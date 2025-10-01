package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// --------------------- core.go 脚本的核心在于实现与各个微服务服务的交互与相应的数据格式化 -----------------------------

// getUserPortrait 获取用户画像数据
func getUserPortrait(sessionID string) (UserPortraitDTO, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("http://localhost:%d/user_poritrait/get/%s", Config.Server.UserPortrait, sessionID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return UserPortraitDTO{}, err
	}
	req.Header.Set("Authorization", "Bearer "+Config.Auth.Token)

	resp, err := client.Do(req)
	if err != nil {
		return UserPortraitDTO{}, err
	}
	defer resp.Body.Close()

	var result QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return UserPortraitDTO{}, fmt.Errorf("failed to decode response: %v", err)
	}
	if result.Code != 0 {
		return UserPortraitDTO{}, fmt.Errorf("user_portrait error: %s", result.Msg)
	}

	// 先 marshal result.Data 再 unmarshal 到 DTO
	rawData, err := json.Marshal(result.Data)
	if err != nil {
		return UserPortraitDTO{}, fmt.Errorf("marshal data error: %v", err)
	}

	var userResp struct {
		UserPortrait UserPortraitDTO `json:"UserPortrait"`
	}
	if err := json.Unmarshal(rawData, &userResp); err != nil {
		return UserPortraitDTO{}, fmt.Errorf("unmarshal user portrait error: %v", err)
	}

	return userResp.UserPortrait, nil
}

// getTopicSummary 获取主题归纳数据
func getTopicSummary(sessionID, query string) ([]TopicSummaryDTO, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("http://localhost:%d/topic_summary/search/%s?q=%s",
		Config.Server.TopicSummary,
		sessionID,
		url.QueryEscape(query),
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+Config.Auth.Token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %v", err)
	}

	// 通用响应结构
	var result struct {
		Code int             `json:"code"`
		Msg  string          `json:"msg"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("decode response failed: %v, body: %s", err, string(bodyBytes))
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("topic_summary error: %s, body: %s", result.Msg, string(bodyBytes))
	}

	// 解析 Data
	var topicData []struct {
		Topic   string `json:"Topic"`
		Content string `json:"Content"`
	}

	if len(result.Data) > 0 && string(result.Data) != "{}" {
		if err := json.Unmarshal(result.Data, &topicData); err != nil {
			return nil, fmt.Errorf("decode topic data failed: %v, data: %s", err, string(result.Data))
		}
	}

	// 聚合同一话题的内容
	topicMap := make(map[string][]string)
	for _, item := range topicData {
		topicMap[item.Topic] = append(topicMap[item.Topic], item.Content)
	}

	dtoList := make([]TopicSummaryDTO, 0, len(topicMap))
	for topic, contents := range topicMap {
		dtoList = append(dtoList, TopicSummaryDTO{
			Topic:   topic,
			Content: contents,
		})
	}

	Info("success getTopicSummary: %d", len(dtoList))
	return dtoList, nil
}

// getChatEvents 获取关键事件数据
func getChatEvents(sessionID string) (ChatEventsDTO, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/chat_event/get/%s", Config.Server.ChatEvent, sessionID), nil)
	if err != nil {
		return ChatEventsDTO{}, err
	}
	req.Header.Set("Authorization", "Bearer "+Config.Auth.Token)

	resp, err := client.Do(req)
	if err != nil {
		return ChatEventsDTO{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ChatEventsDTO{}, fmt.Errorf("chat_event request failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Code int        `json:"code"`
		Msg  string     `json:"msg"`
		Data *EventData `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ChatEventsDTO{}, fmt.Errorf("failed to decode response: %v", err)
	}
	if result.Code != 0 {
		return ChatEventsDTO{}, fmt.Errorf("chat_event error: %s", result.Msg)
	}

	// 转换成 DTO
	dto := ChatEventsDTO{}
	if result.Data != nil {
		dto.Completed = make([]string, len(result.Data.Completed))
		for i, item := range result.Data.Completed {
			dto.Completed[i] = item.Event
		}
		dto.Todo = make([]string, len(result.Data.Todo))
		for i, item := range result.Data.Todo {
			dto.Todo[i] = item.Event
		}
	}

	return dto, nil
}

// getSessionMessages 获取会话消息数据
func getSessionMessages(sessionID string) (SessionMessagesDTO, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	urlStr := fmt.Sprintf("http://localhost:%d/session_messages/get/%s", Config.Server.SessionMessages, sessionID)
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return SessionMessagesDTO{}, err
	}
	req.Header.Set("Authorization", "Bearer "+Config.Auth.Token)

	resp, err := client.Do(req)
	if err != nil {
		return SessionMessagesDTO{}, err
	}
	defer resp.Body.Close()

	var result QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return SessionMessagesDTO{}, fmt.Errorf("failed to decode response: %v", err)
	}
	if result.Code != 0 {
		return SessionMessagesDTO{}, fmt.Errorf("session_messages error: %s", result.Msg)
	}

	var data SessionMessagesDTO
	if err := json.Unmarshal(result.Data, &data); err != nil {
		return SessionMessagesDTO{}, fmt.Errorf("failed to unmarshal session_messages data: %v", err)
	}

	return data, nil
}

// deleteUserPortrait 删除用户画像数据
func deleteUserPortrait(sessionID string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("http://localhost:%d/user_poritrait/delete/%s", Config.Server.UserPortrait, sessionID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+Config.Auth.Token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}
	//Info("deleteUserPortrait result: %v", result)
	if result.Code != 0 {
		return fmt.Errorf("user_poritrait delete error: %s", result.Msg)
	}

	return nil
}

// deleteTopicSummary 删除主题归纳数据
func deleteTopicSummary(sessionID string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("http://localhost:%d/topic_summary/delete/%s", Config.Server.TopicSummary, sessionID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+Config.Auth.Token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}
	if result.Code != 0 {
		return fmt.Errorf("topic_summary delete error: %s", result.Msg)
	}

	return nil
}

// deleteChatEvents 删除关键事件数据
func deleteChatEvents(sessionID string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("http://localhost:%d/chat_event/delete/%s", Config.Server.ChatEvent, sessionID)
	Info("request delete chat_event in url: %s", url)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+Config.Auth.Token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}
	if result.Code != 0 {
		return fmt.Errorf("chat_event delete error: %s", result.Msg)
	}

	return nil
}

// deleteSessionMessages 删除会话消息数据
func deleteSessionMessages(sessionID string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("http://localhost:%d/session_messages/delete/%s", Config.Server.SessionMessages, sessionID)
	//Info("request delete session_messages in url: %s", url)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+Config.Auth.Token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}
	if result.Code != 0 {
		return fmt.Errorf("session_messages delete error: %s", result.Msg)
	}

	return nil
}
