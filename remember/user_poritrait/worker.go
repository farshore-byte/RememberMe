package user_poritrait

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
	DBClient     *UserClient
	Template     *UserProfileTemplate
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
				"*Task failed after %d retries!*\nTaskID: %s\nSessionID: %s\nMessages: %+v\nLastError: %v",
				MaxRetry, msg.TaskID, msg.SessionID, msg.Messages, err,
			)
			go SendFeishuMsgAsync(alertText)

			log.Printf("⚠️ Task dropped after %d retries, task_id=%s, last error: %v", MaxRetry, msg.TaskID, err)
		}
	}
}

// --------------------- 用户画像处理逻辑 -------------------
func (w *Worker) processMessages(msg *QueueMessage) error {
	// 1. 消息转文本
	messagesStr := MessagesToText(msg.Messages)

	// 2. 查询当前用户画像
	userPortrait, err := w.DBClient.GetUserPortrait(msg.SessionID)
	if err != nil {
		return fmt.Errorf("获取用户画像失败: %w", err)
	}

	// 3. 用户画像转 JSON
	userProfileStr, err := Struct2JSON(userPortrait.UserPortrait)
	if err != nil {
		return fmt.Errorf("转换用户画像失败: %w", err)
	}

	// 4. 构造系统提示词
	dynamicVars := &UserProfileDynamicVars{
		CurrentUserPortrait: userProfileStr,
		MesssagesStr:        messagesStr,
		CurrentTime:         FormatTimestamp(msg.Timestamp),
	}
	systemPrompt, err := w.Template.BuildPrompt(dynamicVars)
	if err != nil {
		return fmt.Errorf("生成系统提示词失败: %w", err)
	}

	// 5. 执行模型
	req := &ExecuteRequest{
		Client:       &OpenAIClient,
		SystemPrompt: systemPrompt,
		Query:        User_query,
		Model:        LLMModel,
	}
	result, err := Execute(req)
	if err != nil {
		return fmt.Errorf("执行模型失败: %w", err)
	}
	log.Println("✅生成用户画像成功:", result.JSON)

	// 6. 合并画像
	mergedPortrait := updateUserPortrait(userPortrait.UserPortrait, result.JSON)

	// 7. 写入数据库
	createdAt := userPortrait.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	newUserPortrait := &UserPortrait{
		ID:           userPortrait.ID,
		SessionID:    msg.SessionID,
		UserPortrait: mergedPortrait,
		CreatedAt:    createdAt,
		UpdatedAt:    time.Now().UTC(),
	}

	if err := w.DBClient.UploadUserPortrait(newUserPortrait); err != nil {
		return fmt.Errorf("更新用户画像失败: %w", err)
	}

	return nil
}

// ------------------ 辅助函数 合并更新两个用户画像 -------------------
func updateUserPortrait(oldPortrait, newPortrait map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range oldPortrait {
		if _, ok := UserPortraitOneStepFields[k]; !ok {
			log.Printf("Warning: current Portrait exist field %s is not in allowed fields, ignore it", k)
			continue
		}
		// 直接复制已有且被允许的字段
		merged[k] = v
	}

	for field, value := range newPortrait {
		if _, ok := UserPortraitOneStepFields[field]; !ok {
			log.Printf("Warning: field %s is not in allowed fields, ignore it", field)
			continue
		}

		// 确保是 map[string]interface{}
		newMap, ok := value.(map[string]interface{})
		if !ok {
			log.Printf("Warning: field %s has unknown type, ignore it", field)
			continue
		}

		oldMap, ok := merged[field].(map[string]interface{})
		if !ok {
			oldMap = make(map[string]interface{})
		}

		for subField, subValue := range newMap {
			// 替换已有字段或新增字段
			oldMap[subField] = subValue
		}
		merged[field] = oldMap
	}

	return merged
}
