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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 连接 MongoDB
	fmt.Println("MongoDB 索引检查与创建脚本 - Go版本")
	fmt.Println("=" + repeatString("=", 49))
	mongoCfg := config.Config.MongoDB
	uri :=mongoCfg.URI
	databaseName := mongoCfg.DB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db := client.Database(databaseName)
	coll := db.Collection("topic_summary")

	// 打印当前索引
	fmt.Println("当前索引列表：")
	cursor, err := coll.Indexes().List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	var indexes []bson.M
	cursor.All(ctx, &indexes)
	for _, idx := range indexes {
		fmt.Println(idx)
	}

	// 删除已有的文本索引（如果有）
	for _, idx := range indexes {
		if name, ok := idx["name"].(string); ok && name != "_id_" {
			fmt.Printf("删除索引: %s\n", name)
			_, err := coll.Indexes().DropOne(ctx, name)
			if err != nil {
				log.Printf("⚠️ 删除索引失败: %v\n", err)
			}
		}
	}

	// 创建复合文本索引
	indexModel := mongo.IndexModel{
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
			}).
			SetDefaultLanguage("english"),
	}

	name, err := coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Fatalf("创建文本索引失败: %v", err)
	}
	fmt.Println("创建的文本索引名称:", name)

	// 插入测试数据
	testData := []interface{}{
		bson.M{"session_id": "role_test", "topic": "health", "keywords": "herb medicine", "content": "Herbal remedies are natural medicines."},
		bson.M{"session_id": "role_test", "topic": "food", "keywords": "spice", "content": "This dish uses many herbs."},
		bson.M{"session_id": "role_test", "topic": "technology", "keywords": "AI ML", "content": "Artificial intelligence is evolving."},
	}

	_, err = coll.InsertMany(ctx, testData)
	if err != nil {
		log.Fatalf("插入测试数据失败: %v", err)
	}
	fmt.Println("✅ 插入测试数据成功")

	// 文本搜索
	searchTerm := "herb ddd"
	filter := bson.D{
		{"$text", bson.D{{"$search", searchTerm}}},
		{"session_id", "role_test"},
	}

	opts := options.Find().
		SetProjection(bson.M{"score": bson.M{"$meta": "textScore"}}).
		SetSort(bson.M{"score": bson.M{"$meta": "textScore"}})

	cursor, err = coll.Find(ctx, filter, opts)
	if err != nil {
		log.Fatalf("文本搜索失败: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("文本搜索结果 (%d 条):\n", len(results))
	for i, res := range results {
		b, _ := bson.MarshalExtJSON(res, true, false)
		fmt.Printf("%d: %s\n", i+1, string(b))
	}
}
