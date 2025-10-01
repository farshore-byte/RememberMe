package user_poritrait

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type UserClient struct {
	Collection *mongo.Collection
}

var DBClient *UserClient

func init() {
	DBClient = NewUserClient()
}
func NewUserClient() *UserClient {
	return &UserClient{
		Collection: MongoDB.Collection(DB_NAME),
	}
}

/*
type UserPortrait struct {
	ID           string                 `bson:"_id"`           // 用户 ID
	SessionID    string                 `bson:"session_id"`    // 会话 ID
	UserPortrait map[string]interface{} `bson:"user_portrait"` // 用户画像
	CreatedAt    time.Time              `bson:"created_at"`    // 创建时间
	UpdatedAt    time.Time              `bson:"updated_at"`    // 最后更新时间
}
*/

// 插入新的用户画像，如果用户已存在，则更新用户画像，如果不存在，则插入新的用户画像
func (uc *UserClient) UploadUserPortrait(portrait *UserPortrait) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	portrait.UpdatedAt = time.Now().UTC()
	if portrait.CreatedAt.IsZero() {
		portrait.CreatedAt = time.Now().UTC()
	}

	filter := bson.M{"session_id": portrait.SessionID}
	update := bson.M{
		"$set": bson.M{
			"user_portrait": portrait.UserPortrait,
			"updated_at":    portrait.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"_id":        portrait.ID,
			"session_id": portrait.SessionID,
			"created_at": portrait.CreatedAt,
		},
	}

	opts := options.Update().SetUpsert(true)

	_, err := uc.Collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// 获取用户画像
func (uc *UserClient) GetUserPortrait(sessionID string) (*UserPortrait, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"session_id": sessionID}
	var result UserPortrait

	err := uc.Collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// 找不到记录时返回空 UserPortrait
			return &UserPortrait{
				ID:           GenerateUUID(),
				SessionID:    sessionID,
				UserPortrait: make(map[string]interface{}),
				CreatedAt:    time.Now().UTC(),
				UpdatedAt:    time.Now().UTC(),
			}, nil
		}
		return nil, err
	}

	return &result, nil
}

// DeleteUserPortrait 删除指定 sessionID 的用户画像
func (uc *UserClient) DeleteUserPortrait(sessionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"session_id": sessionID}

	res, err := uc.Collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		// 可选：没有删除到记录也返回提示
		return nil
	}

	return nil
}
