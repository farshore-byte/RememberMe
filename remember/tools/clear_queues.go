package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"remember/config"

	"github.com/redis/go-redis/v9"
)

// 所有微服务的Redis队列名称
var queueNames = []string{
	"remember:main:queue",           // 主服务
	"remember:chat_event:queue",     // 关键事件
	"remember:topic_summary:queue",  // 主题归纳
	"remember:user_poritrait:queue", // 用户画像
	// session_messages 服务没有使用队列
}

func main() {
	log.Println("🚀 开始清空Remember系统的Redis任务队列...")

	// 获取Redis配置
	redisConfig := config.Config.Redis
	log.Printf("📡 连接到Redis: %s:%d (DB: %d)", redisConfig.Host, redisConfig.Port, redisConfig.DB)

	// 创建Redis客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("❌ Redis连接失败: %v", err)
	}
	log.Println("✅ Redis连接成功")

	// 清空所有队列
	var totalCleared int64 = 0
	for _, queueName := range queueNames {
		cleared, err := clearQueue(ctx, rdb, queueName)
		if err != nil {
			log.Printf("⚠️  清空队列 %s 时出错: %v", queueName, err)
		} else {
			log.Printf("✅ 队列 %s 已清空，删除了 %d 个任务", queueName, cleared)
			totalCleared += cleared
		}
	}

	// 关闭Redis连接
	if err := rdb.Close(); err != nil {
		log.Printf("⚠️  关闭Redis连接时出错: %v", err)
	}

	log.Printf("🎉 清空完成！总共清空了 %d 个任务", totalCleared)
	log.Println("📊 清空的队列统计:")
	for _, queueName := range queueNames {
		log.Printf("   - %s", queueName)
	}
}

// clearQueue 清空指定的Redis队列
func clearQueue(ctx context.Context, rdb *redis.Client, queueName string) (int64, error) {
	// 获取队列长度
	queueLen, err := rdb.LLen(ctx, queueName).Result()
	if err != nil {
		return 0, fmt.Errorf("获取队列长度失败: %v", err)
	}

	if queueLen == 0 {
		log.Printf("ℹ️  队列 %s 已经是空的", queueName)
		return 0, nil
	}

	log.Printf("📊 队列 %s 当前有 %d 个任务", queueName, queueLen)

	// 清空队列 - 使用DEL命令删除整个列表
	result, err := rdb.Del(ctx, queueName).Result()
	if err != nil {
		return 0, fmt.Errorf("删除队列失败: %v", err)
	}

	return result, nil
}
