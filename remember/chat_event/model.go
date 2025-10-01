package chat_event

import "time"

type ChatEvent struct {
	ID            string    `bson:"_id"`            // 唯一主键
	SessionID     string    `bson:"session_id"`     // 会话 ID
	CreatedAt     time.Time `bson:"created_at"`     // 创建时间
	Event         string    `bson:"event"`          // 事件描述
	ExecutionTime time.Time `bson:"execution_time"` // 根据语境推断的事件发生时间
	EventType     int       `bson:"event_type"`     // 1: 已完成事件, 2: 代办事项
}
