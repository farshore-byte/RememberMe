package chat_event

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

type EventClient struct {
	Collection *mongo.Collection
}

var DBClient *EventClient

func init() {
	DBClient = NewEventClient()
}

// -------- 初始化函数 ----------
func NewEventClient() *EventClient {
	return &EventClient{
		Collection: MongoDB.Collection(DB_NAME),
	}
}

// 插入一条事件，事件字段：
/*
type ChatEvent struct {
	ID            string    `bson:"_id"`            // 唯一主键，可用 UUID 或 user_id_role_id + 时间戳
	SessionID     string    `bson:"session_id"`     // 会话 ID
	CreatedAt     time.Time `bson:"created_at"`     // 创建时间
	Event         string    `bson:"event"`          // 事件描述
	ExecutionTime time.Time `bson:"execution_time"` // 根据语境腿短的事件发生时间
	EventType     int       `bson:"event_type"`     // 1: 已完成事件, 2: 代办事项
}
*/

// 插入一条事件

func (ec *EventClient) UploadChatEvent(event *ChatEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := ec.Collection.InsertOne(ctx, event)
	if err != nil {
		return err
	}
	return nil
}

// 查询指定会话的时间列表，在event_type==1 中取最近的五个事件，从event_type==2 中取最新的五个事件

func (ec *EventClient) GetSessionEvents(sessionID string) (map[string][]*ChatEvent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 聚合管道
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"session_id": sessionID}}},
		{{
			Key: "$facet",
			Value: bson.M{
				"completed": []bson.M{
					{"$match": bson.M{"event_type": 1}},     // 已完成事件
					{"$sort": bson.M{"execution_time": -1}}, // 按执行时间倒序
					{"$limit": 5},                           // 只取最新 5 条
				},
				"todo": []bson.M{
					{"$match": bson.M{"event_type": 2}}, // 待办事件
					{"$sort": bson.M{"execution_time": -1}},
					{"$limit": 5},
				},
			},
		}},
	}

	cur, err := ec.Collection.Aggregate(ctx, pipeline, options.Aggregate())
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	// 直接解析 facet 返回的结果
	var results []struct {
		Completed []*ChatEvent `bson:"completed"`
		Todo      []*ChatEvent `bson:"todo"`
	}

	if err := cur.All(ctx, &results); err != nil {
		return nil, err
	}

	// 封装成 map 返回
	res := make(map[string][]*ChatEvent)
	if len(results) > 0 {
		res["completed"] = results[0].Completed
		res["todo"] = results[0].Todo
	} else {
		res["completed"] = []*ChatEvent{}
		res["todo"] = []*ChatEvent{}
	}

	return res, nil
}

// DeleteSessionEvents 删除指定 sessionID 的所有 ChatEvent
func (ec *EventClient) DeleteSessionEvents(sessionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"session_id": sessionID}
	res, err := ec.Collection.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}

	// 可选：打印删除数量
	if res != nil {
		log.Printf("✅ 删除 %d 条 ChatEvent, session_id=%s", res.DeletedCount, sessionID)
	}
	return nil
}
