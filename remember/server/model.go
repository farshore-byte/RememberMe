package server

import "encoding/json"

// ----------------------  upload 接口 -------------------------

// Message 聊天消息
type Message struct {
	Role    string `json:"role" bson:"role"`
	Content string `json:"content" bson:"content"`
}

// UploadRequest 上传接口请求体
type UploadRequest struct {
	SessionID string    `json:"session_id"`
	UserID    string    `json:"user_id"`
	RoleID    string    `json:"role_id"`
	GroupID   string    `json:"group_id"`
	Messages  []Message `json:"messages"`
}

// UploadResponse 上传接口响应
type UploadResponse struct {
	Code int         `json:"code"` // 0 成功, -1 失败
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

//-------------------------- query 查询  接口 ----------

// QueryRequest 查询接口请求体
type QueryRequest struct {
	SessionID string `json:"session_id"`
	Query     string `json:"query"` // 可选
}

// QueryResponse 微服务查询接口响应-通用
type QueryResponse struct {
	Code int             `json:"code"` // 0 成功, -1 失败
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

// server 多个微服务结果汇总，format之后的查询响应
type FormResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		UserPortrait    UserPortraitDTO   `json:"user_portrait"`
		TopicSummary    []TopicSummaryDTO `json:"topic_summary"`
		ChatEvents      ChatEventsDTO     `json:"chat_events"`
		SessionMessages []Message         `json:"session_messages"`
		CurrentTime     string            `json:"current_time"`
	} `json:"data"`
}

// 关键事件查询接口返回数据结构体
type EventItem struct {
	ID            string `json:"ID"`
	SessionID     string `json:"SessionID"`
	CreatedAt     string `json:"CreatedAt"`
	Event         string `json:"Event"`
	ExecutionTime string `json:"ExecutionTime"`
	EventType     int    `json:"EventType"`
}

type EventData struct {
	Completed []EventItem `json:"completed"`
	Todo      []EventItem `json:"todo"`
}

// 结构化后的聊天事件 DTO
type ChatEventsDTO struct {
	Completed []string `json:"completed"`
	Todo      []string `json:"todo"` // 只保留事件描述字符串
}

// 消息查询响应结构体，且无需格式化，同为DTO
type SessionMessagesDTO struct {
	Messages []Message `json:"messages"`
}

// 话题归纳查询响应结构体
type TopicSummaryRaw struct {
	Topic   string `json:"topic"`
	Content string `json:"content"`
	// Keywords, ID 等字段可省略
}



// 话题归纳数据，apply接口专用
type TopicSummaryData []TopicSummaryRaw


// 话题归纳core返回type，apply 接口专用
type TopicSummaryResult struct {
	TopicList []string
	Data      TopicSummaryData `json:"data"`
}


// 结构化后的主题归纳 DTO
type TopicSummaryDTO struct {
	Topic   string   `json:"topic"`
	Content []string `json:"content"` // 按话题归类的列表
}

// 用户画像查询结构体
type UserPortraitRaw struct {
	UserPortrait json.RawMessage `json:"UserPortrait"`
}

// 结构化后的用户画像 DTO
type UserPortraitDTO map[string]interface{}

// -------------------------   apply memory 接口 -------------------------------------
// 请求体结构
type ApplyRequest struct {
	SessionID  string `json:"session_id,omitempty"`
	UserID     string `json:"user_id,omitempty"`
	RoleID     string `json:"role_id,omitempty"`
	GroupID    string `json:"group_id,omitempty"`
	RolePrompt string `json:"role_prompt"` // 角色设定提示词
	Query      string `json:"query"`  // 可选，可为空
}

// 响应体结构
type ApplyResponse struct {
	Code int       `json:"code"`
	Msg  string    `json:"msg"`
	Data ApplyData `json:"data"`
}

type ApplyData struct {
	SystemPrompt string    `json:"system_prompt"` // 不加json，直接使用字段名作为json的key
	Messages     []Message `json:"messages"`      // json key : messages
}

// -------------------------   delete 接口 -------------------------------------
// DeleteRequest 删除接口请求体
type DeleteRequest struct {
	SessionID string `json:"session_id"`
}

// DeleteResponse 删除接口响应
type DeleteResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// DeleteResult 删除结果详情
type DeleteResult struct {
	ServiceName string `json:"service_name"`
	Success     bool   `json:"success"`
	Message     string `json:"message"`
}
