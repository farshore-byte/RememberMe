package topic_summary

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

// Worker 消费队列消息
type Worker struct {
	Queue        *QueueClient
	StopCh       chan struct{}
	PollInterval time.Duration
	DBClient     *TopicClient
	Template     *TopicTemplat
}

// NewWorker 创建 Worker
func NewWorker(interval time.Duration) *Worker {
	return &Worker{
		Queue:        MessageQueue,
		StopCh:       make(chan struct{}),
		PollInterval: interval,
		DBClient:     DBClient, // 全局 DBClient
		Template:     NewTopicSummaryTemplate(),
	}
}

// Start 启动 Worker
func (w *Worker) Start() {
	go func() {
		for {
			select {
			case <-w.StopCh:
				fmt.Println("Worker stopped")
				return
			default:
				w.processNext()
				time.Sleep(w.PollInterval)
			}
		}
	}()
}

// Stop 停止 Worker
func (w *Worker) Stop() {
	close(w.StopCh)
}

// processNext 处理队列中的下一条消息
func (w *Worker) processNext() {
	ctx := context.Background()
	msg, err := w.Queue.Dequeue(ctx)
	if err != nil {
		if err.Error() != "redis: nil" { // 队列为空
			log.Printf("Error dequeue message: %v\n", err)
		}
		return
	}

	log.Printf("Processing session_id=%s, task_id=%s, retry=%d", msg.SessionID, msg.TaskID, msg.Retry)

	if err := w.processTopicSummary(msg); err != nil {
		log.Printf("❌ Task failed, session_id=%s, task_id=%s, retry=%d, err=%v", msg.SessionID, msg.TaskID, msg.Retry, err)

		// 判断是否需要重试
		if msg.Retry < MaxRetry {
			msg.Retry++
			if _, enqueueErr := w.Queue.Enqueue(ctx, *msg); enqueueErr != nil {
				log.Printf("❌ Re-enqueue failed, task_id=%s, err=%v", msg.TaskID, enqueueErr)
			} else {
				log.Printf("🔁 Task re-enqueued, session_id=%s, task_id=%s, retry=%d", msg.SessionID, msg.TaskID, msg.Retry)
			}
		} else {
			// 超过重试次数，发送飞书报警
			alertText := fmt.Sprintf(
				"*Topic summary task failed after %d retries!*\nTaskID: %s\nSessionID: %s\nLastError: %v",
				MaxRetry, msg.TaskID, msg.SessionID, err,
			)
			go SendFeishuMsgAsync(alertText)

			log.Printf("⚠️ Task dropped after %d retries, task_id=%s, last error: %v", MaxRetry, msg.TaskID, err)
		}
	}
}

// processTopicSummary 处理话题摘要逻辑
func (w *Worker) processTopicSummary(msg *QueueMessage) error {
	// 1. 消息转文本
	messagesStr := MessagesToText(msg.Messages)

	// 2. 构造系统提示词
	dynamicVars := map[string]string{
		"messages_str": messagesStr,
	}
	systemPrompt, err := w.Template.BuildPrompt(dynamicVars)
	if err != nil {
		return fmt.Errorf("%s 生成系统提示词失败: %w", SERVER_NAME, err)
	}

	// 3. 执行模型
	req := &ExecuteRequest{
		SystemPrompt: systemPrompt,
		Query:        User_query,
		Model:        Config.LLM.ModelID,
		Client:       &OpenAIClient,
	}
	result, err := Execute(req)
	if err != nil {
		return fmt.Errorf("%s 执行模型失败: %w", SERVER_NAME, err)
	}

	log.Printf("✅ 生成话题摘要成功, task_id=%s", msg.TaskID)
	if len(result.JSON) == 0 {
		log.Printf("%s task_id=%s 话题摘要为空, 跳过上传", SERVER_NAME, msg.TaskID)
		return nil
	}

	// 4. 上传到数据库
	if err := w.DBClient.UploadTopicSummary(context.Background(), msg, result.JSON); err != nil {
		return fmt.Errorf("%s 上传话题摘要失败: %w", SERVER_NAME, err)
	}

	log.Printf("✅ 上传话题摘要成功, session_id=%s, task_id=%s, topics_count=%d",
		msg.SessionID, msg.TaskID, len(result.JSON))
	return nil
}

// 初始化 OpenAI 客户端
var OpenAIClient openai.Client

func init() {
	InitLLM()
}

// InitLLM 初始化 OpenAI Client
func InitLLM() {
	OpenAIClient = openai.NewClient(
		option.WithAPIKey(Config.LLM.APIKey),
		option.WithBaseURL(Config.LLM.BaseURL),
	)
}
