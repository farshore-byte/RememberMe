package topic_summary

import "time"

// TopicRecord 话题表
type TopicRecord struct {
	ID        string    `bson:"_id"`        // 唯一主键
	SessionID string    `bson:"session_id"` // 会话 ID
	Topic     string    `bson:"topic"`      // 话题名称
	Content   string    `bson:"content"`    // 内容
	Keywords  []string  `bson:"keywords"`   // RAKE算法提取的关键词（用户后续分析）
	CreatedAt time.Time `bson:"created_at"` // 创建时间
	UpdatedAt time.Time `bson:"updated_at"` // 更新时间

	// MongoDB 文本搜索返回的分数 (只在投影时使用 $meta: "textScore" 才会有值)
	Score float64 `bson:"score,omitempty" json:"score,omitempty"`
}

// 话题统计表
type TopicInfo struct {
	SessionID string `bson:"session_id"` // 会话 ID
	//UserID       string        `bson:"user_id"`       // 用户ID
	//RoleID       string        `bson:"role_id"`       // 角色ID
	//GroupID      string        `bson:"group_id"`      // 群/会话ID
	TopicCount   int           `bson:"topic_count"`   // 当前会话的话题总数
	ActiveTopics []ActiveTopic `bson:"active_topics"` // 最近活跃的 N 个话题
	UpdatedAt    time.Time     `bson:"updated_at"`    // 最近一次更新
}

// ActiveTopic 用于表示活跃话题
type ActiveTopic struct {
	Topic      string    `bson:"topic"`       // 话题名称
	LastActive time.Time `bson:"last_active"` // 最近活跃时间
}
