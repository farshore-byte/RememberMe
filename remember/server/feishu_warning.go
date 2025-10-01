package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const PrefixWord = "🚨记忆服务\n[调度器]:\n"

type FeishuMsg struct {
	MsgType string      `json:"msg_type"`
	Content interface{} `json:"content"`
}

type TextContent struct {
	Text string `json:"text"`
}

// 异步发送飞书消息
func SendFeishuMsgAsync(text string) chan error {
	result := make(chan error, 1) // 带缓冲，避免阻塞
	go func() {
		err := sendFeishuMsg(text)
		result <- err
		close(result)
	}()
	return result
}

// 内部发送函数
func sendFeishuMsg(text string) error {
	webhook := Config.Feishu.Webhook
	if webhook == "" {
		return fmt.Errorf("feishu webhook not configured")
	}

	payload := FeishuMsg{
		MsgType: "text",
		Content: TextContent{
			Text: PrefixWord + text,
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json marshal error: %v", err)
	}

	resp, err := http.Post(webhook, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("http post error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("feishu returned status %d", resp.StatusCode)
	}

	log.Println("✅飞书消息发送成功")
	return nil
}

