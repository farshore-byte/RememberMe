package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	// 测试 mark_task 接口
	client := &http.Client{Timeout: 10 * time.Second}

	// 准备测试数据
	markReq := map[string]interface{}{
		"session_id": "test_session_123",
		"task_index": 2, // 聊天事件任务索引
		"task_id":    "test_task_456",
	}

	jsonData, err := json.Marshal(markReq)
	if err != nil {
		fmt.Printf("JSON序列化失败: %v\n", err)
		return
	}

	fmt.Printf("发送请求到 mark_task 接口...\n")
	fmt.Printf("请求数据: %s\n", string(jsonData))

	// 发送请求到 session_messages 服务的 mark_task 接口
