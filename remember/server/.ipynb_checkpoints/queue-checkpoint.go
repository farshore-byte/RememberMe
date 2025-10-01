package server

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// QueueMessage 队列消息结构
type QueueMessage struct {
	TaskID    string    `json:"task_id"`
	SessionID string    `json:"session_id"`
	Messages  []Message `json:"messages"`
	Timestamp int64     `json:"timestamp"`
	Retry     int       `json:"retry"`
}

// QueueClient 封装队列操作
type QueueClient struct {
	RedisClient *redis.Client
	QueueName   string
}

var MessageQueue *QueueClient

func init() {
	MessageQueue = NewQueueClient()

}

// NewQueueClient 创建 QueueClient
func NewQueueClient() *QueueClient {
	return &QueueClient{
		RedisClient: RedisClient,
		QueueName:   QUEUE_NAME,
	}
}

// Enqueue 入队列
func (q *QueueClient) Enqueue(ctx context.Context, msg QueueMessage) (string, error) {
	// 如果 TaskID 为空，则生成
	if msg.TaskID == "" {
		msg.TaskID = GenerateUUID()
	}
	// 初始化重试次数
	if msg.Retry == 0 {
		msg.Retry = 0
	}
	// 设置时间戳
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().UTC().Unix()
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return msg.TaskID, err
	}

	if err := q.RedisClient.RPush(ctx, q.QueueName, data).Err(); err != nil {
		return msg.TaskID, err
	}

	Info("%s Enqueued message for session_id=%s, task_id=%s, retry=%d", SERVER_NAME, msg.SessionID, msg.TaskID, msg.Retry)
	return msg.TaskID, nil
}

// Dequeue 出队列
func (q *QueueClient) Dequeue(ctx context.Context) (*QueueMessage, error) {
	result, err := q.RedisClient.LPop(ctx, q.QueueName).Result()
	if err != nil {
		return nil, err
	}

	var msg QueueMessage
	if err := json.Unmarshal([]byte(result), &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

// Length 获取队列长度
func (q *QueueClient) Length() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	length, err := q.RedisClient.LLen(ctx, q.QueueName).Result()
	if err != nil {
		return 0, err
	}
	return length, nil
}

// DeleteBySession 删除队列中指定 sessionID 的消息
func (q *QueueClient) DeleteBySession(ctx context.Context, sessionID string) error {
	// 获取整个队列
	messages, err := q.RedisClient.LRange(ctx, q.QueueName, 0, -1).Result()
	if err != nil {
		return err
	}

	for _, raw := range messages {
		var msg QueueMessage
		if err := json.Unmarshal([]byte(raw), &msg); err != nil {
			continue // 出错就跳过
		}

		if msg.SessionID == sessionID {
			// 从队列中删除该条消息
			if err := q.RedisClient.LRem(ctx, q.QueueName, 1, raw).Err(); err != nil {
				return err
			}
		}
	}

	return nil
}
