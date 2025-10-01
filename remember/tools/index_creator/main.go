package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	fmt.Println("MongoDB 索引创建脚本 - Go版本")
	fmt.Println("=" + repeatString("=", 49))

	// 连接参数配置 - 请根据您的环境修改
	uri := "mongodb://localhost:27017/"
	databaseName := "remember"

	// 如果有认证，使用以下格式：
	// uri = "mongodb://username:password@localhost:27017/"

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 连接到MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("❌ 连接MongoDB失败: %v", err)
	}
	defer client.Disconnect(ctx)

	// 检查连接
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("❌ 无法ping通MongoDB: %v", err)
	}

	db := client.Database(databaseName)

	fmt.Println("开始创建MongoDB索引...")

	// ==================== chat_event 集合索引 ====================
	fmt.Println("\n=== 创建 chat_event 集合索引 ===")

	// 主要查询索引
	chatEventIndex1 := mongo.IndexModel{
		Keys: bson.D{
			{Key: "session_id", Value: 1},
			{Key: "event_type", Value: 1},
			{Key: "execution_time", Value: -1},
		},
		Options: options.Index().SetName("session_event_time_idx").SetBackground(true),
	}

	// 会话ID索引
	chatEventIndex2 := mongo.IndexModel{
		Keys:    bson.D{{Key: "session_id", Value: 1}},
		Options: options.Index().SetName("session_id_idx").SetBackground(true),
	}

	_, err = db.Collection("chat_event").Indexes().CreateMany(ctx, []mongo.IndexModel{chatEventIndex1, chatEventIndex2})
	if err != nil {
		fmt.Printf("⚠️  chat_event索引创建失败: %v\n", err)
	} else {
		fmt.Println("✅ 创建 chat_event 索引")
	}

	// ==================== session_messages 集合索引 ====================
	fmt.Println("\n=== 创建 session_messages 集合索引 ===")

	sessionMessagesIndexes := []mongo.IndexModel{
		// 主要查询索引
		{
			Keys: bson.D{
				{Key: "session_id", Value: 1},
				{Key: "created_at", Value: 1},
			},
			Options: options.Index().SetName("session_created_idx").SetBackground(true),
		},
		// 清理逻辑索引
		{
			Keys: bson.D{
				{Key: "session_id", Value: 1},
				{Key: "task1_id", Value: 1},
				{Key: "task2_id", Value: 1},
				{Key: "task3_id", Value: 1},
			},
			Options: options.Index().SetName("session_tasks_idx").SetBackground(true),
		},
		// 任务标记查询索引
		{
			Keys:    bson.D{{Key: "session_id", Value: 1}, {Key: "task1_id", Value: 1}},
			Options: options.Index().SetName("session_task1_idx").SetBackground(true),
		},
		{
			Keys:    bson.D{{Key: "session_id", Value: 1}, {Key: "task2_id", Value: 1}},
			Options: options.Index().SetName("session_task2_idx").SetBackground(true),
		},
		{
			Keys:    bson.D{{Key: "session_id", Value: 1}, {Key: "task3_id", Value: 1}},
			Options: options.Index().SetName("session_task3_idx").SetBackground(true),
		},
	}

	_, err = db.Collection("session_messages").Indexes().CreateMany(ctx, sessionMessagesIndexes)
	if err != nil {
		fmt.Printf("⚠️  session_messages索引创建失败: %v\n", err)
	} else {
		fmt.Println("✅ 创建 session_messages 索引")
	}

	// ==================== topic_summary 集合索引 ====================
	fmt.Println("\n=== 创建 topic_summary 集合索引 ===")

	topicSummaryIndexes := []mongo.IndexModel{
		// 复合索引 - 支持精确查询
		{
			Keys:    bson.D{{Key: "session_id", Value: 1}, {Key: "topic", Value: 1}},
			Options: options.Index().SetName("session_topic_idx").SetBackground(true),
		},
		// 会话ID索引
		{
			Keys:    bson.D{{Key: "session_id", Value: 1}},
			Options: options.Index().SetName("session_id_idx").SetBackground(true),
		},
	}

	_, err = db.Collection("topic_summary").Indexes().CreateMany(ctx, topicSummaryIndexes)
	if err != nil {
		fmt.Printf("⚠️  topic_summary索引创建失败: %v\n", err)
	} else {
		fmt.Println("✅ 创建 topic_summary 索引")
	}

	// 全文搜索索引（需要单独创建）
	fmt.Println("创建 topic_summary 全文搜索索引...")
	textIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "topic", Value: "text"},
			{Key: "content", Value: "text"},
			{Key: "keywords", Value: "text"},
		},
		Options: options.Index().
			SetName("text_search_idx").
			SetBackground(true).
			SetWeights(bson.D{
				{Key: "topic", Value: 10},
				{Key: "keywords", Value: 8},
				{Key: "content", Value: 5},
			}),
	}

	_, err = db.Collection("topic_summary").Indexes().CreateOne(ctx, textIndexModel)
	if err != nil {
		fmt.Printf("⚠️  topic_summary全文搜索索引创建失败: %v\n", err)
	} else {
		fmt.Println("✅ 创建 topic_summary 全文搜索倒排索引")
		fmt.Println("   - topic字段权重: 10")
		fmt.Println("   - keywords字段权重: 8")
		fmt.Println("   - content字段权重: 5")
	}

	// ==================== topic_info 集合索引 ====================
	fmt.Println("\n=== 创建 topic_info 集合索引 ===")

	topicInfoIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "session_id", Value: 1}},
		Options: options.Index().SetName("session_id_idx").SetBackground(true),
	}

	_, err = db.Collection("topic_info").Indexes().CreateOne(ctx, topicInfoIndex)
	if err != nil {
		fmt.Printf("⚠️  topic_info索引创建失败: %v\n", err)
	} else {
		fmt.Println("✅ 创建 topic_info 索引")
	}

	// ==================== user_portrait 集合索引 ====================
	fmt.Println("\n=== 创建 user_portrait 集合索引 ===")

	userPortraitIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "session_id", Value: 1}},
		Options: options.Index().SetName("session_id_idx").SetBackground(true),
	}

	_, err = db.Collection("user_portrait").Indexes().CreateOne(ctx, userPortraitIndex)
	if err != nil {
		fmt.Printf("⚠️  user_portrait索引创建失败: %v\n", err)
	} else {
		fmt.Println("✅ 创建 user_portrait 索引")
	}

	// ==================== 索引创建完成 ====================
	fmt.Println("\n🎉 所有索引创建完成！")

	// 显示索引创建统计
	fmt.Println("\n📊 索引创建统计:")
	collections := []string{"chat_event", "session_messages", "topic_summary", "topic_info", "user_portrait"}
	for _, collectionName := range collections {
		cursor, err := db.Collection(collectionName).Indexes().List(ctx)
		if err != nil {
			fmt.Printf("  %s: 无法获取索引信息\n", collectionName)
			continue
		}

		var indexes []bson.M
		if err := cursor.All(ctx, &indexes); err != nil {
			fmt.Printf("  %s: 无法解析索引信息\n", collectionName)
			continue
		}

		fmt.Printf("  %s: %d 个索引\n", collectionName, len(indexes))
	}

	fmt.Println("\n💡 使用建议:")
	fmt.Println("1. 在生产环境建议在业务低峰期执行索引创建")
	fmt.Println("2. 使用 background=true 选项避免阻塞数据库操作")
	fmt.Println("3. 定期监控索引使用情况和性能")
	fmt.Println("4. 根据实际查询模式调整索引策略")
}

func repeatString(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
