package topic_summary

import (
	"context"
	"log"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TopicClient struct {
	SummaryCollection *mongo.Collection // 话题记录集合
	InfoCollection    *mongo.Collection // 会话信息集合
}

var DBClient *TopicClient

func init() {
	DBClient = NewTopicClient()
}

func NewTopicClient() *TopicClient {
	return &TopicClient{
		SummaryCollection: MongoDB.Collection(DB_NAME),
		InfoCollection:    MongoDB.Collection(DB_NAME_2),
	}
}

// UploadTopicSummary 上传话题摘要
func (tc *TopicClient) UploadTopicSummary(ctx context.Context, msg *QueueMessage, json_data map[string]interface{}) error {
	topics := make([]string, 0, len(json_data))

	// 处理每个话题
	for topic, content := range json_data {
		topics = append(topics, topic)

		// 提取关键词
		keywords := ExtractKeywords(content.(string))

		// 构造记录
		record := TopicRecord{
			ID:        GenerateUUID(),
			SessionID: msg.SessionID,
			Topic:     topic,
			Content:   content.(string),
			Keywords:  keywords,
			CreatedAt: FormatTimestamp(msg.Timestamp),
			UpdatedAt: FormatTimestamp(msg.Timestamp),
		}

		// 插入记录
		_, err := tc.SummaryCollection.InsertOne(ctx, record)
		if err != nil {
			return err
		}
	}

	// 更新会话信息
	return tc.updateTopicInfo(ctx, msg.SessionID, FormatTimestamp(msg.Timestamp), topics)
}

// updateSessionInfo 更新会话信息
func (tc *TopicClient) updateTopicInfo(ctx context.Context, sessionID string, createdAt time.Time, topics []string) error {
	topicInfo := TopicInfo{}
	filter := bson.M{"session_id": sessionID}

	err := tc.InfoCollection.FindOne(ctx, filter).Decode(&topicInfo)
	if err != nil {
		// 没有记录时创建新记录
		if err == mongo.ErrNoDocuments {
			topicInfo = TopicInfo{
				SessionID:    sessionID,
				TopicCount:   0,
				ActiveTopics: make([]ActiveTopic, 0),
				UpdatedAt:    time.Now().UTC(),
			}
			_, err = tc.InfoCollection.InsertOne(ctx, topicInfo)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// 统计话题总数
	count, err := tc.SummaryCollection.CountDocuments(ctx, bson.M{"session_id": sessionID})
	if err != nil {
		return err
	}
	topicInfo.TopicCount = int(count)

	// 检查是否超过最大话题数量限制，如果超过则删除最旧的话题
	if topicInfo.TopicCount > MAX_TOPIC_COUNT {
		topicsToDelete := topicInfo.TopicCount - MAX_TOPIC_COUNT
		err := tc.deleteOldestTopics(ctx, sessionID, topicsToDelete)
		if err != nil {
			log.Printf("⚠️ 删除最旧话题失败: %v", err)
		} else {
			log.Printf("✅ 删除 %d 个最旧话题，保持话题数量不超过 %d", topicsToDelete, MAX_TOPIC_COUNT)
			// 重新统计话题总数
			count, err = tc.SummaryCollection.CountDocuments(ctx, bson.M{"session_id": sessionID})
			if err != nil {
				return err
			}
			topicInfo.TopicCount = int(count)
		}
	}

	// 更新活跃话题列表
	maxCount := ActivateTopicCount(topicInfo.TopicCount)
	activeTopics := topicInfo.ActiveTopics

	for _, topic := range topics {
		found := false
		for i := range activeTopics {
			if activeTopics[i].Topic == topic {
				activeTopics[i].LastActive = createdAt
				found = true
				break
			}
		}
		if !found {
			activeTopics = append(activeTopics, ActiveTopic{
				Topic:      topic,
				LastActive: createdAt,
			})
		}
	}

	// 排序（最旧 -> 最新）
	sort.Slice(activeTopics, func(i, j int) bool {
		return activeTopics[i].LastActive.Before(activeTopics[j].LastActive)
	})

	// 截断
	if len(activeTopics) > maxCount {
		activeTopics = activeTopics[len(activeTopics)-maxCount:]
	}

	topicInfo.ActiveTopics = activeTopics
	topicInfo.UpdatedAt = createdAt

	// 更新数据库
	filter = bson.M{"session_id": sessionID}
	update := bson.M{"$set": topicInfo}
	_, err = tc.InfoCollection.UpdateOne(ctx, filter, update)
	return err
}

// GetTopicSummary 查询话题摘要

// （活跃话题全取，非活跃话题使用关键词搜索）
func (tc *TopicClient) GetTopicSummary(
	ctx context.Context,
	sessionID string,
	query string,
	activeTopics []string,
) ([]TopicRecord, error) {

	allResults := make([]TopicRecord, 0)
	seen := make(map[string]bool) // 去重（按 _id 唯一标识）

	// --- 第一步：取活跃话题 ---
	if len(activeTopics) > 0 {
		activeFilter := bson.M{
			"session_id": sessionID,
			"topic":      bson.M{"$in": activeTopics},
		}

		activeResults, err := tc.findTopics(ctx, activeFilter, nil)
		if err != nil {
			log.Printf("⚠️ GetTopicSummary 第一步取活跃话题失败: %v", err)
		} else {
			for _, r := range activeResults {
				if !seen[r.ID] {
					seen[r.ID] = true
					allResults = append(allResults, r)
				}
			}
		}
	}

	// --- 第二步：非活跃话题 + 关键词搜索 ---
	inactiveResults, err := tc.SearchInactiveTopics(ctx, sessionID, query, activeTopics)
	if err != nil {
		log.Printf("⚠️ GetTopicSummary 第二步搜索失败: %v", err)
		inactiveResults = []TopicRecord{}
	}

	for _, r := range inactiveResults {
		if !seen[r.ID] {
			seen[r.ID] = true
			allResults = append(allResults, r)
		}
	}

	return allResults, nil
}


// --- 公共查询函数 ---
// 封装 Find + Close + Decode，避免重复代码
func (tc *TopicClient) findTopics(ctx context.Context, filter interface{}, opts *options.FindOptions) ([]TopicRecord, error) {
	cursor, err := tc.SummaryCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []TopicRecord
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// DeleteTopic 删除指定话题
/*
func (tc *TopicClient) DeleteTopic(ctx context.Context, sessionID string, topic string) error {
	filter := bson.M{"session_id": sessionID, "topic": topic}

	res, err := tc.SummaryCollection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if res.DeletedCount > 0 {
		log.Printf("✅ 删除话题: session_id=%s, topic=%s", sessionID, topic)
	}

	return nil
}
*/

// DeleteSessionTopics 删除指定会话的所有话题
func (tc *TopicClient) DeleteSessionTopics(ctx context.Context, sessionID string) error {
	filter := bson.M{"session_id": sessionID}

	res, err := tc.SummaryCollection.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}

	if res.DeletedCount > 0 {
		log.Printf("✅ 删除 %d 条话题记录, session_id=%s", res.DeletedCount, sessionID)
	}

	// 同时删除会话信息
	return tc.DeleteSessionInfo(ctx, sessionID)
}

// DeleteSessionInfo 删除指定会话的信息记录
func (tc *TopicClient) DeleteSessionInfo(ctx context.Context, sessionID string) error {
	filter := bson.M{"session_id": sessionID}
	_, err := tc.InfoCollection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	log.Printf("✅ 删除会话信息记录, session_id=%s", sessionID)
	return nil
}

// deleteOldestTopics 删除最旧的话题（按话题分组，删除最旧的话题记录）
func (tc *TopicClient) deleteOldestTopics(ctx context.Context, sessionID string, count int) error {
	if count <= 0 {
		return nil
	}

	// 按话题分组，找出每个话题的最早记录
	pipeline := []bson.M{
		{"$match": bson.M{"session_id": sessionID}},
		{"$group": bson.M{
			"_id":            "$topic",
			"min_created_at": bson.M{"$min": "$created_at"},
			"records":        bson.M{"$push": "$$ROOT"},
		}},
		{"$sort": bson.M{"min_created_at": 1}}, // 按最早创建时间排序
		{"$limit": int64(count)},
	}

	cursor, err := tc.SummaryCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var topicGroups []struct {
		Topic        string        `bson:"_id"`
		MinCreatedAt time.Time     `bson:"min_created_at"`
		Records      []TopicRecord `bson:"records"`
	}

	if err := cursor.All(ctx, &topicGroups); err != nil {
		return err
	}

	// 删除这些话题的所有记录
	for _, group := range topicGroups {
		_, err := tc.SummaryCollection.DeleteMany(ctx, bson.M{
			"session_id": sessionID,
			"topic":      group.Topic,
		})
		if err != nil {
			return err
		}
		log.Printf("✅ 删除话题 '%s' 的所有记录 (%d 条)", group.Topic, len(group.Records))
	}

	return nil
}

// ActivateTopicCount 计算活跃话题数量
func ActivateTopicCount(count int) int {
	if count <= 10 {
		return count
	}
	return int(float64(count-10)*0.5) + 10
}


// SearchInactiveTopics 搜索非活跃话题，使用 $text 查询关键词
func (tc *TopicClient) SearchInactiveTopics(
	ctx context.Context,
	sessionID string,
	query string,
	activeTopics []string,
) ([]TopicRecord, error) {

	if query == "" {
		return nil, nil
	}

	keywords := ExtractKeywords(query)
	if len(keywords) == 0 {
		return nil, nil
	}

	searchQuery := strings.Join(keywords, " ")

	// --- 搜索条件 ---
	filter := bson.D{
		{"$text", bson.D{{"$search", searchQuery}}},
		{"session_id", sessionID},
	}

	if len(activeTopics) > 0 {
		filter = append(filter, bson.E{Key: "topic", Value: bson.M{"$nin": activeTopics}})
	}

	// --- 搜索选项：按文本评分排序并返回评分 ---
	opts := options.Find().
		SetProjection(bson.M{"score": bson.M{"$meta": "textScore"}}).
		SetSort(bson.M{"score": bson.M{"$meta": "textScore"}})

	// 执行查询
	results, err := tc.findTopics(ctx, filter, opts)
	if err != nil {
		log.Printf("⚠️ SearchInactiveTopics 搜索失败: sessionID=%s query=%s err=%v", sessionID, query, err)
		return nil, err
	}

	// 分数过滤（阈值可调）
	filtered := make([]TopicRecord, 0)
	for _, r := range results {
		if r.Score >= 0.7 {
			filtered = append(filtered, r)
		}
	}

	return filtered, nil
}
