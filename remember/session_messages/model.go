package session_messages

import "time"

// MemoryMessage 短期消息表
type MemoryMessage struct {
	ID               string    `bson:"_id,omitempty"`     // MongoDB 唯一主键，可用 UUID 或自增
	SessionID        string    `bson:"session_id"`        // 会话 ID
	UserContent      string    `bson:"user_content"`      // 用户输入
	AssistantContent string    `bson:"assistant_content"` // 助手回复
	CreatedAt        time.Time `bson:"created_at"`        // 创建时间
	MessagesID       string    `bson:"messages_id"`       // 消息轮次ID
	//-------------- taskN 的设计是为了区分不同任务的完成情况，有task_id则说明该任务正在进行中或者已完成
	Task1  string `bson:"task1_id"` // 任务1 用户画像
	Task2  string `bson:"task2_id"` // 任务2 关键事件
	Task3  string `bson:"task3_id"` // 任务3 主题归纳
	Task4  string `bson:"task4_id"` // 任务4 预留位
	Status int    `bson:"status"`   // 状态  1: 已完成   0: 待处理  -1: 失败
}

// Message 消息结构
type Message struct {
	Role    string `json:"role" bson:"role"`
	Content string `json:"content" bson:"content"`
}
