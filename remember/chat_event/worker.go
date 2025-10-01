package chat_event

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Worker 消费队列消息
type Worker struct {
	Queue        *QueueClient
	StopCh       chan struct{}
	PollInterval time.Duration
	DBClient     *EventClient
	Template     *ChatEventTemplate
}

// NewWorker 创建 Worker
func NewWorker(interval time.Duration) *Worker {
	return &Worker{
		Queue:        MessageQueue,
		StopCh:       make(chan struct{}),
		PollInterval: interval,
		DBClient:     DBClient, // 全局 DBClient
		Template:     Template, // 全局 Template
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
// ----------------------------------------------------------

//-----------------------------------------------------

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

	if err := w.processMessages(msg); err != nil {
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
			// 超过重试次数，发送飞书报警，并附上最新报错
			alertText := fmt.Sprintf(
				"*Task failed after %d retries!*\nTaskID: %s\nSessionID: %s\nConversations: %+v\nLastError: %v",
				MaxRetry, msg.TaskID, msg.SessionID, msg.Conversations, err,
			)
			go SendFeishuMsgAsync(alertText)

			log.Printf("⚠️ Task dropped after %d retries, task_id=%s, last error: %v", MaxRetry, msg.TaskID, err)
		}
	}
}

// processMessages 处理关键事件逻辑
func (w *Worker) processMessages(msg *QueueMessage) error {
	// 1. 将对话对转换为文本，保留时间戳信息
	conversationsStr := ConversationsToText(msg.Conversations)

	// 2. 构造系统提示词
	currentTime := FormatTimestamp(msg.Timestamp)
	dynamicVars := &ChatEventDynamicVars{
		MessagesStr: conversationsStr,
		CurrentTime: currentTime,
	}
	systemPrompt, err := w.Template.BuildPrompt(dynamicVars)
	if err != nil {
		return fmt.Errorf("%s 生成系统提示词失败: %w", SERVER_NAME, err)
	}

	// 3. 执行模型
	req := &ExecuteRequest{
		Client:       &OpenAIClient,
		SystemPrompt: systemPrompt,
		Query:        User_query,
		Model:        LLMModel,
	}
	result, err := Execute(req)
	if err != nil {
		return fmt.Errorf("%s 执行模型失败: %w", SERVER_NAME, err)
	}

	log.Println("✅生成关键事件成功:", result.JSON)
	if len(result.JSON) == 0 {
		log.Printf("%s task_id=%s 关键事件为空, 跳过事件上传", SERVER_NAME, msg.TaskID)
		return nil
	}
	rawEvents := result.JSON

	// 5. 根据时间分类事件并上传到数据库
	now := time.Now().UTC()
	for tsStr, eventContent := range rawEvents {
		t, err := ParseTimestamp(tsStr)
		if err != nil {
			log.Printf("%s ⚠️ 无法解析时间 %s, 忽略该事件", SERVER_NAME, tsStr)
			continue
		}
		eventType := 1 // 过去事件
		fmt.Println("t:", t.Format(time.RFC3339Nano))
		fmt.Println("now:", now.Format(time.RFC3339Nano))
		if t.After(now) {
			log.Printf("%s 当前时间 %s 检测到 [%s] 事件时间 %s 未来, 该时间标记为未来事件", SERVER_NAME, now, eventContent, tsStr)
			eventType = 2 // 未来事件
		}else {
			log.Printf("%s 当前时间 %s 检测到 [%s] 事件时间 %s 过去, 该时间标记为过去事件", SERVER_NAME, now, eventContent, tsStr)
		}
		chatEvent := ChatEvent{
			ID:            GenerateUUID(),
			SessionID:     msg.SessionID,
			CreatedAt:     t,
			Event:         fmt.Sprintf("%v", eventContent),
			ExecutionTime: time.Now().UTC(),
			EventType:     eventType,
		}

		// 上传到数据库
		if err := w.DBClient.UploadChatEvent(&chatEvent); err != nil {
			log.Printf("❌ 上传 ChatEvent 失败, task_id=%s, err=%v", msg.TaskID, err)
			// 重试上传
			return fmt.Errorf("%s 上传 ChatEvent 失败: %w", SERVER_NAME, err)

		} else {
			log.Printf("✅ 上传 ChatEvent 成功, task_id=%s, time=%s", msg.TaskID, tsStr)
		}
	}

	return nil
}
