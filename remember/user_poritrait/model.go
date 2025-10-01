package user_poritrait

import "time"

// UserPortrait 用户画像记录
type UserPortrait struct {
	ID           string                 `bson:"_id"`           // 用户 ID
	SessionID    string                 `bson:"session_id"`    // 会话 ID
	UserPortrait map[string]interface{} `bson:"user_portrait"` // 用户画像
	CreatedAt    time.Time              `bson:"created_at"`    // 创建时间
	UpdatedAt    time.Time              `bson:"updated_at"`    // 最后更新时间
}

// 一级字段约束
var UserPortraitOneStepFields = map[string]struct{}{
	"basic_information":  {},
	"interest_topics":    {},
	"sexual_orientation": {},
}
