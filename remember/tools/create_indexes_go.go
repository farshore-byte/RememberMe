package main

import (
	"context"
	"fmt"
	"log"
	"time"
	"remember/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	fmt.Println("MongoDB 索引检查与创建脚本 - Go版本")
	fmt.Println("=" + repeatString("=", 49))
	mongoCfg := config.Config.MongoDB
	uri :=mongoCfg.URI
	databaseName := mongoCfg.DB

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("❌ 连接MongoDB失败: %v", err)
	}
	defer client.Disconnect(ctx)

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("❌ 无法ping通MongoDB: %v", err)
	}

	db := client.Database(databaseName)
	//打印索引
	fmt.Println("数据库名称:", databaseName)


	fmt.Println("开始检查并创建MongoDB索引...")

	// ==================== 通用函数 ====================
	checkAndCreateIndex := func(collName string, idx mongo.IndexModel) {
		coll := db.Collection(collName)

		cursor, err := coll.Indexes().List(ctx)
		if err != nil {
			fmt.Printf("⚠️ 获取 %s 索引列表失败: %v\n", collName, err)
			return
		}

		var indexes []bson.M
		if err := cursor.All(ctx, &indexes); err != nil {
			fmt.Printf("⚠️ 解析 %s 索引列表失败: %v\n", collName, err)
			return
		}

		for _, existing := range indexes {
			if name, ok := existing["name"].(string); ok && name == *idx.Options.Name {
				fmt.Printf("✅ %s 已存在索引: %s\n", collName, name)
				return
			}
		}

		_, err = coll.Indexes().CreateOne(ctx, idx)
		if err != nil {
			fmt.Printf("⚠️ 创建 %s 索引失败: %v\n", collName, err)
		} else {
			fmt.Printf("✅ 创建 %s 索引: %s\n", collName, *idx.Options.Name)
		}
	}

	// ==================== chat_event 集合索引 ====================
	fmt.Println("\n=== chat_event 集合索引 ===")
	chatEventIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "session_id", Value: 1},
				{Key: "event_type", Value: 1},
				{Key: "execution_time", Value: -1},
			},
			Options: options.Index().SetName("session_event_time_idx").SetBackground(true),
		},
		{
			Keys:    bson.D{{Key: "session_id", Value: 1}},
			Options: options.Index().SetName("session_id_idx").SetBackground(true),
		},
	}
	for _, idx := range chatEventIndexes {
		checkAndCreateIndex("chat_event", idx)
	}
	// ==================== chat_event TTL 索引 ====================
	fmt.Println("\n=== chat_event TTL 索引 ===")
	chatEventTTLIndex := mongo.IndexModel{
	Keys: bson.D{{Key: "created_at", Value: 1}},
	Options: options.Index().
		SetName("created_at_ttl_idx").
		SetExpireAfterSeconds(7 * 24 * 3600), // 7 天
		}
	checkAndCreateIndex("chat_event", chatEventTTLIndex)



	// ==================== session_messages 集合索引 ====================
	fmt.Println("\n=== session_messages 集合索引 ===")
	sessionMessagesIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "session_id", Value: 1}, {Key: "created_at", Value: 1}},
			Options: options.Index().SetName("session_created_idx").SetBackground(true),
		},
		{
			Keys:    bson.D{{Key: "session_id", Value: 1}, {Key: "task1_id", Value: 1}, {Key: "task2_id", Value: 1}, {Key: "task3_id", Value: 1}},
			Options: options.Index().SetName("session_tasks_idx").SetBackground(true),
		},
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
	for _, idx := range sessionMessagesIndexes {
		checkAndCreateIndex("session_messages", idx)
	}

	// ==================== topic_summary 集合索引 ====================
	fmt.Println("\n=== topic_summary 集合索引 ===")
	// 查看索引
	// 打印 topic_summary 集合当前所有索引
printIndexes := func(coll *mongo.Collection) {
	fmt.Println("\n📜 topic_summary 当前索引列表:")
	cursor, err := coll.Indexes().List(context.Background())
	if err != nil {
		fmt.Println("⚠️ 获取索引失败:", err)
		return
	}
	var indexes []bson.M
	if err := cursor.All(context.Background(), &indexes); err != nil {
		fmt.Println("⚠️ 解析索引失败:", err)
		return
	}
	for _, idx := range indexes {
		fmt.Println(idx)
	}
}

printIndexes(db.Collection("topic_summary"))
	topicSummaryIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "session_id", Value: 1}, {Key: "topic", Value: 1}},
			Options: options.Index().SetName("session_topic_idx").SetBackground(true),
		},
		{
			Keys:    bson.D{{Key: "session_id", Value: 1}},
			Options: options.Index().SetName("session_id_idx").SetBackground(true),
		},
		{
			Keys:    bson.D{{Key: "session_id", Value: 1}, {Key: "created_at", Value: 1}},
			Options: options.Index().SetName("session_created_at_idx").SetBackground(true),
		},
		{
			Keys:    bson.D{{Key: "created_at", Value: 1}},
			Options: options.Index().SetName("created_at_idx").SetBackground(true),
		},
	}
	for _, idx := range topicSummaryIndexes {
		checkAndCreateIndex("topic_summary", idx)
	}

	// 全文搜索索引
	fmt.Println("检查 topic_summary 全文搜索索引...")
	// 检查是否存在 text_search_idx，存在就删除
	dropIndexIfExists(ctx, db.Collection("topic_summary"), "text_search_idx")
	// 检查是否存在 session_text_search_idx，存在就删除
	dropIndexIfExists(ctx, db.Collection("topic_summary"), "session_text_search_idx")
	//删除text字段
	dropIndexIfExists(ctx, db.Collection("topic_summary"), "text")
	// 创建全文搜索索引
	sessionTextIndexModel := mongo.IndexModel{
    Keys: bson.D{
        {Key: "topic", Value: "text"},
        {Key: "keywords", Value: "text"},
        {Key: "content", Value: "text"},
    },
    Options: options.Index().
        SetName("session_text_search_idx").
        SetBackground(true).
        SetWeights(bson.D{
            {Key: "topic", Value: 10},
            {Key: "keywords", Value: 8},
            {Key: "content", Value: 5},
        }),
	}
	checkAndCreateIndex("topic_summary", sessionTextIndexModel)


	// ==================== topic_info 集合索引 ====================
	fmt.Println("\n=== topic_info 集合索引 ===")
	topicInfoIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "session_id", Value: 1}},
		Options: options.Index().SetName("session_id_idx").SetBackground(true),
	}
	checkAndCreateIndex("topic_info", topicInfoIndex)

	// ==================== user_portrait 集合索引 ====================
	fmt.Println("\n=== user_portrait 集合索引 ===")
	userPortraitIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "session_id", Value: 1}},
		Options: options.Index().SetName("session_id_idx").SetBackground(true),
	}
	checkAndCreateIndex("user_portrait", userPortraitIndex)

	fmt.Println("\n🎉 所有索引检查完成！")
}

// repeatString 重复字符
func repeatString(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

// 辅助函数，检查删除索引
// 删除索引（如果存在）
func dropIndexIfExists(ctx context.Context, coll *mongo.Collection, indexName string) {
	cursor, err := coll.Indexes().List(ctx)
	if err != nil {
		fmt.Printf("⚠️ 获取 %s 索引列表失败: %v\n", coll.Name(), err)
		return
	}

	var indexes []bson.M
	if err := cursor.All(ctx, &indexes); err != nil {
		fmt.Printf("⚠️ 解析 %s 索引列表失败: %v\n", coll.Name(), err)
		return
	}

	for _, idx := range indexes {
		if name, ok := idx["name"].(string); ok && name == indexName {
			_, err := coll.Indexes().DropOne(ctx, indexName)
			if err != nil {
				fmt.Printf("⚠️ 删除 %s 索引 %s 失败: %v\n", coll.Name(), indexName, err)
			} else {
				fmt.Printf("✅ 删除 %s 索引: %s\n", coll.Name(), indexName)
			}
			return
		}
	}
}
