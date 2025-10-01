package session_messages

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MessageClient struct {
	Collection *mongo.Collection
}

var DBClient *MessageClient

func init() {
	DBClient = NewMessageClient()
}

func NewMessageClient() *MessageClient {
	return &MessageClient{
		Collection: MongoDB.Collection(DB_NAME),
	}
}

// InsertMessage 插入新的消息记录
func (mc *MessageClient) InsertMessage(message *MemoryMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now().UTC()
	}

	if message.ID == "" {
		message.ID = GenerateUUID()
	}

	_, err := mc.Collection.InsertOne(ctx, message)

	// --------------------- 打印日志 ---------------------
	//Info(fmt.Sprintf("%s insert message %s sccuess", SERVER_NAME, message.ID))

	return err
}

// GetMessagesBySessionID 根据 session_id 查询消息列表
func (mc *MessageClient) GetMessagesBySessionID(sessionID string) ([]MemoryMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"session_id": sessionID}
	opts := options.Find().SetSort(bson.D{
		{Key: "created_at", Value: 1}, // 按创建时间升序
	})

	cursor, err := mc.Collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []MemoryMessage
	if err = cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	//----------------- 打印日志 ----------------------
	// Info(fmt.Sprintf("%s get %d messages from session %s", SERVER_NAME, len(messages), sessionID))

	return messages, nil
}

// UpdateMessageStatus 更新消息状态
func (mc *MessageClient) UpdateMessageStatus(messageID string, status int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": messageID}
	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}

	_, err := mc.Collection.UpdateOne(ctx, filter, update)
	return err
}

// DeleteMessagesBySessionID 删除指定 session_id 的所有消息
func (mc *MessageClient) DeleteMessagesBySessionID(sessionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"session_id": sessionID}

	deleteResult, err := mc.Collection.DeleteMany(ctx, filter)

	// -------------------------  打印日志 ----------------------------------
	Warn(fmt.Sprintf("%s delete %d messages from session %s", SERVER_NAME, deleteResult.DeletedCount, sessionID))

	return err
}

//  清理逻辑是清理走完流程的消息，但是不保证所有任务处理都成功，这个逻辑考虑到微服务的分离，因此后续要回调函数

// --------------------  清理逻辑 ------------------------

// clearSessionMessages 清理指定 session 下 task1、task2、task3 全部完成的消息（目前只有用到这三个任务， 因此只判断这三个）
//
// -----------------------------  新增：最近消息保护：最近 project_messages_count 条消息必定保留 --------------------------------
// clearSessionMessages 清理指定 session 下 task1、task2、task3 全部完成的消息
func (mc *MessageClient) clearSessionMessages(sessionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 首先获取当前session的总消息数
	totalCount, err := mc.CountMessagesBySessionID(sessionID)
	if err != nil {
		return err
	}

	// 过滤条件：task1_id、task2_id、task3_id 都不为空
	filter := bson.M{
		"session_id": sessionID,
		"task1_id":   bson.M{"$ne": ""},
		"task2_id":   bson.M{"$ne": ""},
		"task3_id":   bson.M{"$ne": ""},
	}

	// 查询符合条件的消息，按时间排序（假设有 timestamp 字段，越大越新）
	cursor, err := mc.Collection.Find(ctx, filter, options.Find().SetSort(bson.M{"timestamp": 1}))
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var filteredMessages []bson.M
	if err := cursor.All(ctx, &filteredMessages); err != nil {
		return err
	}


	filteredCount := len(filteredMessages)


	if filteredCount <= PROJECT_MESSAGES_COUNT {
		// 如果过滤出的消息数量小于等于要保留的数量，则不删除任何消息
		//Info(fmt.Sprintf(" ♻️ %s no delete, only %d messages found (<= %d)", SERVER_NAME, filteredCount, keepCount))
		Info("♻️ filter messages count <= keep count, no need to delete.")
		return nil
	}

	
	// 如果总消息数 - 过滤出的消息数 >= PROJECT_MESSAGES_COUNT，直接删除所有过滤出的消息
	if int(totalCount)-filteredCount >= PROJECT_MESSAGES_COUNT {
		// 删除所有符合条件的消息
		Info(fmt.Sprintf("all masked messages can be deleted."))
		deleteResult, err := mc.Collection.DeleteMany(ctx, filter)
		if err != nil {
			return err
		}
		Info(fmt.Sprintf(" ♻️  %s delete %d messages (all 3 tasks done, keep last %d) from session %s", SERVER_NAME, deleteResult.DeletedCount, PROJECT_MESSAGES_COUNT, sessionID))
		return nil
	}
	keepCount := int(PROJECT_MESSAGES_COUNT) - (int(totalCount) - filteredCount)
	//keepCount取值 [0, PROJECT_MESSAGES_COUNT]


	// 从过滤出的消息中，keepCount 条消息保留，其余的删除

	toDelete := filteredMessages[:filteredCount-keepCount]
	var ids []interface{}
	for _, msg := range toDelete {
		if id, ok := msg["_id"]; ok {
			ids = append(ids, id)
		}
	}

	if len(ids) == 0 {
		return nil
	}

	deleteFilter := bson.M{"_id": bson.M{"$in": ids}}
	deleteResult, err := mc.Collection.DeleteMany(ctx, deleteFilter)
	if err != nil {
		return err
	}

	Info(fmt.Sprintf(" ♻️ %s delete %d messages (all 3 tasks done, keep last %d) from session %s", SERVER_NAME, deleteResult.DeletedCount, PROJECT_MESSAGES_COUNT, sessionID))
	return nil
}

/*
// clearSessionMessages 指定sessionID, 清理status == 1 的消息

func (mc *MessageClient) clearSessionMessages(sessionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{"session_id": sessionID, "status": 1}
	deleteResult, err := mc.Collection.DeleteMany(ctx, filter)

	//-------------------------    打印日志 --------------------------------------
	Info(fmt.Sprintf(" %s delete %d messages from session %s", SERVER_NAME, deleteResult.DeletedCount, sessionID))

	if err != nil {
		return err
	}
	return nil
}


*/

// 统计指定sessionID下的消息数量
func (mc *MessageClient) CountMessagesBySessionID(sessionID string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"session_id": sessionID}
	count, err := mc.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// 查找指定 session 下 taskN_id 为空的消息，并标记为指定 taskID，取消息和标记消息合并
func (mc *MessageClient) FindAndMarkMessagesWithoutTaskID(sessionID string, taskIndex int, taskID string) ([]MemoryMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	taskField := fmt.Sprintf("task%d_id", taskIndex)

	// 查询条件：taskN_id 不存在或为空
	filter := bson.M{
		"session_id": sessionID,
		"$or": []bson.M{
			{taskField: bson.M{"$exists": false}},
			{taskField: ""},
		},
	}

	// 先查出需要更新的消息
	var messages []MemoryMessage
	cursor, err := mc.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &messages); err != nil {
		return nil, err
	}
	cursor.Close(ctx)

	// 如果没有符合条件的消息，直接返回
	if len(messages) == 0 {
		return []MemoryMessage{}, nil
	}

	// 更新这些消息的 taskN_id
	update := bson.M{"$set": bson.M{taskField: taskID}}
	_, err = mc.Collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	// 更新返回结果里的 taskN_id
	for i := range messages {
		switch taskIndex {
		case 1:
			messages[i].Task1 = taskID
		case 2:
			messages[i].Task2 = taskID
		case 3:
			messages[i].Task3 = taskID
		case 4:
			messages[i].Task4 = taskID
		}
	}

	Info(fmt.Sprintf("%s marked %d messages with %s for task%d in session %s",
		SERVER_NAME, len(messages), taskID, taskIndex, sessionID))

	return messages, nil
}
