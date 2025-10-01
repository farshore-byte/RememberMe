package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Worker 消费队列消息
type Worker struct {
	Queue        *QueueClient
	StopCh       chan struct{}
	PollInterval time.Duration
}

// NewWorker 创建 Worker
func NewWorker(interval time.Duration) *Worker {
	return &Worker{
		Queue:        MessageQueue,
		StopCh:       make(chan struct{}),
		PollInterval: interval,
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
	// 处理任务分发
	if err := w.processTaskDistribution(msg); err != nil {
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
				"*main server  task failed after %d retries!*\nTaskID: %s\nSessionID: %s\nLastError: %v",
				MaxRetry, msg.TaskID, msg.SessionID, err,
			)
			go SendFeishuMsgAsync(alertText)

			log.Printf("⚠️ Task dropped after %d retries, task_id=%s, last error: %v", MaxRetry, msg.TaskID, err)
		}
	}
}

// processTaskDistribution 处理任务分发
func (w *Worker) processTaskDistribution(msg *QueueMessage) error {
	// 第一步：上传消息到 session_messages 服务
	if err := uploadToSessionMessages(msg); err != nil {
		return fmt.Errorf("failed to upload to session_messages: %w", err)
	}

	// 第二步：获取当前会话的消息数量
	count, err := getSessionMessagesCount(msg.SessionID)
	if err != nil {
		return fmt.Errorf("failed to get messages count: %w", err)
	}

	log.Printf("Session %s has %d messages", msg.SessionID, count)

	// 第三步：根据消息数量分发任务

	// 关键事件提取任务
	if count%EventRound == 0 {
		if err := triggerChatEventTask(msg.SessionID, msg.TaskID); err != nil {
			return fmt.Errorf("failed to trigger chat event task: %w", err)
		}
		log.Printf("Triggered chat event task for session %s", msg.SessionID)
	}

	// 用户画像任务
	if count%UserRound == 0 {
		if err := triggerUserPortraitTask(msg.SessionID, msg.TaskID); err != nil {
			return fmt.Errorf("failed to trigger user portrait task: %w", err)
		}
		log.Printf("Triggered user portrait task for session %s", msg.SessionID)
	}

	// 主题归纳任务
	if count%TopicRound == 0 {
		if err := triggerTopicSummaryTask(msg.SessionID, msg.TaskID); err != nil {
			return fmt.Errorf("failed to trigger topic summary task: %w", err)
		}
		log.Printf("Triggered topic summary task for session %s", msg.SessionID)
	}

	// 会话清理任务 ，注意这里是大于等于
	if count >= ClearRound {
		if err := cleanSessionMessages(msg.SessionID); err != nil {
			return fmt.Errorf("failed to clean session messages: %w", err)
		}
		log.Printf("Cleaned session messages for session %s", msg.SessionID)
	}

	return nil
}

// uploadToSessionMessages 上传消息到 session_messages 服务
func uploadToSessionMessages(msg *QueueMessage) error {
	client := &http.Client{Timeout: 10 * time.Second}

	uploadReq := map[string]interface{}{
		"session_id": msg.SessionID,
		"messages":   msg.Messages,
		"task_id":    msg.TaskID,
	}

	jsonData, err := json.Marshal(uploadReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/session_messages/upload", Config.Server.SessionMessages), bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+Config.Auth.Token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("session_messages upload failed with status: %d", resp.StatusCode)
	}

	return nil
}

// getSessionMessagesCount 获取会话消息数量
func getSessionMessagesCount(sessionID string) (int, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/session_messages/count/%s", Config.Server.SessionMessages, sessionID), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+Config.Auth.Token)

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get messages count, status: %d", resp.StatusCode)
	}

	var result struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	if result.Code != 0 {
		return 0, fmt.Errorf("failed to get messages count: %s", result.Msg)
	}

	count, ok := result.Data["count"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid count format in response")
	}

	return int(count), nil
}

// triggerChatEventTask 触发聊天事件提取任务
func triggerChatEventTask(sessionID, taskID string) error {
	client := &http.Client{Timeout: 10 * time.Second}

	// 第一步：标记任务状态
	markReq := map[string]interface{}{
		"session_id": sessionID,
		"task_index": 2, // 聊天事件任务索引
		"task_id":    taskID,
	}

	jsonData, err := json.Marshal(markReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/session_messages/mark_task", Config.Server.SessionMessages), bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+Config.Auth.Token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("chat event task mark failed with status: %d", resp.StatusCode)
	}

	// 第二步：从mark_task响应中获取标记后的消息并调用chat_event服务处理
	var markResult struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&markResult); err != nil {
		return fmt.Errorf("failed to decode mark task response: %w", err)
	}

	if markResult.Code != 0 {
		return fmt.Errorf("mark task failed: %s", markResult.Msg)
	}

	messagesData, ok := markResult.Data["messages"]
	if !ok {
		return fmt.Errorf("no messages found in mark task response")
	}

	messages, ok := messagesData.([]interface{})
	if !ok {
		return fmt.Errorf("invalid messages format in mark task response")
	}

	// 如果没有标记到消息，直接返回成功
	if len(messages) == 0 {
		log.Printf("✅ No messages to process for chat event task in session %s", sessionID)
		return nil
	}

	// 将扁平消息列表转换为对话对格式
	conversations, err := convertMessagesToConversations(messages)
	if err != nil {
		return fmt.Errorf("failed to convert messages to conversations: %w", err)
	}

	// 调用chat_event服务的上传接口（使用新的对话对格式）
	eventRequest := map[string]interface{}{
		"session_id":    sessionID,
		"conversations": conversations,
	}

	eventData, err := json.Marshal(eventRequest)
	if err != nil {
		return err
	}

	eventHttpReq, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/chat_event/upload", Config.Server.ChatEvent), bytes.NewReader(eventData))
	if err != nil {
		return err
	}
	eventHttpReq.Header.Set("Content-Type", "application/json")
	eventHttpReq.Header.Set("Authorization", "Bearer "+Config.Auth.Token)

	eventResp, err := client.Do(eventHttpReq)

	if err != nil {
		return err
	}

	// 解析响应
	var Result struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(eventResp.Body).Decode(&Result); err != nil {
		return fmt.Errorf("failed to decode chat event service response: %w", err)
	}

	if Result.Code != 0 {
		return fmt.Errorf("chat event service failed: %s", Result.Msg)
	}

	defer eventResp.Body.Close()

	if eventResp.StatusCode != http.StatusOK {
		return fmt.Errorf("chat event service upload failed with status: %d", eventResp.StatusCode)
	}

	log.Printf("✅ Chat event task triggered successfully for session %s", sessionID)
	return nil
}

// triggerUserPortraitTask 触发用户画像任务
func triggerUserPortraitTask(sessionID, taskID string) error {
	client := &http.Client{Timeout: 10 * time.Second}

	// 第一步：标记任务状态
	markReq := map[string]interface{}{
		"session_id": sessionID,
		"task_index": 1, // 用户画像任务索引
		"task_id":    taskID,
	}

	jsonData, err := json.Marshal(markReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/session_messages/mark_task", Config.Server.SessionMessages), bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+Config.Auth.Token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("user portrait task mark failed with status: %d", resp.StatusCode)
	}

	// 第二步：从mark_task响应中获取标记后的消息并调用user_portrait服务处理
	var markResult struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&markResult); err != nil {
		return fmt.Errorf("failed to decode mark task response: %w", err)
	}

	if markResult.Code != 0 {
		return fmt.Errorf("mark task failed: %s", markResult.Msg)
	}

	messagesData, ok := markResult.Data["messages"]
	if !ok {
		return fmt.Errorf("no messages found in mark task response")
	}

	messages, ok := messagesData.([]interface{})
	if !ok {
		return fmt.Errorf("invalid messages format in mark task response")
	}

	// 如果没有标记到消息，直接返回成功
	if len(messages) == 0 {
		log.Printf("✅ No messages to process for user portrait task in session %s", sessionID)
		return nil
	}

	// 调用user_portrait服务的上传接口
	portraitRequest := map[string]interface{}{
		"session_id": sessionID,
		"messages":   messages,
	}

	portraitData, err := json.Marshal(portraitRequest)
	if err != nil {
		return err
	}

	portraitHttpReq, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/user_poritrait/upload", Config.Server.UserPortrait), bytes.NewReader(portraitData))
	if err != nil {
		return err
	}
	portraitHttpReq.Header.Set("Content-Type", "application/json")
	portraitHttpReq.Header.Set("Authorization", "Bearer "+Config.Auth.Token)

	portraitResp, err := client.Do(portraitHttpReq)
	if err != nil {
		return err
	}
	// 解析响应
	var Result struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(portraitResp.Body).Decode(&Result); err != nil {
		return fmt.Errorf("failed to decode chat event service response: %w", err)
	}

	if Result.Code != 0 {
		return fmt.Errorf("chat event service failed: %s", Result.Msg)
	}

	defer portraitResp.Body.Close()

	if portraitResp.StatusCode != http.StatusOK {
		return fmt.Errorf("user portrait service upload failed with status: %d", portraitResp.StatusCode)
	}

	log.Printf("✅ User portrait task triggered successfully for session %s", sessionID)
	return nil
}

// triggerTopicSummaryTask 触发主题归纳任务
func triggerTopicSummaryTask(sessionID, taskID string) error {
	client := &http.Client{Timeout: 10 * time.Second}

	// 第一步：标记任务状态
	markReq := map[string]interface{}{
		"session_id": sessionID,
		"task_index": 3, // 主题归纳任务索引
		"task_id":    taskID,
	}

	jsonData, err := json.Marshal(markReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/session_messages/mark_task", Config.Server.SessionMessages), bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+Config.Auth.Token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("topic summary task mark failed with status: %d", resp.StatusCode)
	}

	// 第二步：从mark_task响应中获取标记后的消息并调用topic_summary服务处理
	var markResult struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&markResult); err != nil {
		return fmt.Errorf("failed to decode mark task response: %w", err)
	}

	if markResult.Code != 0 {
		return fmt.Errorf("mark task failed: %s", markResult.Msg)
	}

	messagesData, ok := markResult.Data["messages"]
	if !ok {
		return fmt.Errorf("no messages found in mark task response")
	}
	//fmt.Println(messagesData)

	messages, ok := messagesData.([]interface{})
	if !ok {
		return fmt.Errorf("invalid messages format in mark task response")
	}

	// 如果没有标记到消息，直接返回成功
	if len(messages) == 0 {
		log.Printf("✅ No messages to process for topic summary task in session %s", sessionID)
		return nil
	}

	// 调用topic_summary服务的上传接口
	topicRequest := map[string]interface{}{
		"session_id": sessionID,
		"messages":   messages,
	}

	topicData, err := json.Marshal(topicRequest)
	if err != nil {
		return err
	}

	topicHttpReq, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/topic_summary/upload", Config.Server.TopicSummary), bytes.NewReader(topicData))
	if err != nil {
		return err
	}
	topicHttpReq.Header.Set("Content-Type", "application/json")
	topicHttpReq.Header.Set("Authorization", "Bearer "+Config.Auth.Token)

	topicResp, err := client.Do(topicHttpReq)
	if err != nil {
		return err
	}
	// 解析响应
	var Result struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(topicResp.Body).Decode(&Result); err != nil {
		return fmt.Errorf("failed to decode chat event service response: %w", err)
	}

	if Result.Code != 0 {
		return fmt.Errorf("chat event service failed: %s", Result.Msg)
	}
	defer topicResp.Body.Close()

	if topicResp.StatusCode != http.StatusOK {
		return fmt.Errorf("topic summary service upload failed with status: %d", topicResp.StatusCode)
	}

	log.Printf("✅ Topic summary task triggered successfully for session %s", sessionID)
	return nil
}

// convertMessagesToConversations 将扁平消息列表转换为对话对格式
func convertMessagesToConversations(messages []interface{}) ([]map[string]interface{}, error) {
	Info(fmt.Sprintf("convert messages %s to conversations", messages))
	var conversations []map[string]interface{}
	
	// 按顺序处理消息，将连续的 user-assistant 对组合成对话
	var currentConversation []map[string]interface{}
	var lastTimestamp int64 = time.Now().UTC().Unix() // 默认使用当前时间戳
	Info(fmt.Sprintf("mark the current conversation in %s time", time.Now().UTC().Format(time.RFC3339)))
	
	for i, msg := range messages {
		message, ok := msg.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid message format at index %d", i)
		}
		
		role, _ := message["role"].(string)
		content, _ := message["content"].(string)
		
		// 尝试获取时间戳，如果没有则使用默认值
		timestamp := lastTimestamp
		if ts, ok := message["timestamp"].(float64); ok {

			timestamp = int64(ts)
		} else if ts, ok := message["created_at"].(string); ok {
			// 尝试解析时间字符串
			if t, err := time.Parse(time.RFC3339, ts); err == nil {
				timestamp = t.Unix()
			}
		}
		
		msgData := map[string]interface{}{
			"role":    role,
			"content": content,
		}
		
		currentConversation = append(currentConversation, msgData)
		
		// 如果是 assistant 消息，或者到达消息列表末尾，则完成当前对话对
		if role == "assistant" || i == len(messages)-1 {
			if len(currentConversation) > 0 {
				conversation := map[string]interface{}{
					"timestamp": timestamp,
					"messages":  currentConversation,
				}
				conversations = append(conversations, conversation)
				currentConversation = nil
			}
		}
		
		lastTimestamp = timestamp
	}
	
	// 处理剩余的消息（如果有）
	if len(currentConversation) > 0 {
		conversation := map[string]interface{}{
			"timestamp": lastTimestamp,
			"messages":  currentConversation,
		}
		conversations = append(conversations, conversation)
	}
	
	return conversations, nil
}

// cleanSessionMessages 清理会话消息
func cleanSessionMessages(sessionID string) error {
	client := &http.Client{Timeout: 10 * time.Second}

	// 请求体 JSON
	bodyData := map[string]string{"session_id": sessionID}
	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("http://localhost:%d/session_messages/clean", Config.Server.SessionMessages),
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+Config.Auth.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	// 解析响应
	var Result struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&Result); err != nil {
		return fmt.Errorf("failed to decode chat event service response: %w", err)
	}

	if Result.Code != 0 {
		return fmt.Errorf("chat event service failed: %s", Result.Msg)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("session clean failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
