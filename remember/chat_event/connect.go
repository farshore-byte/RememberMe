package chat_event

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MongoClient *mongo.Client
	MongoDB     *mongo.Database
	RedisClient *redis.Client
)

// InitDB 初始化 MongoDB 和 Redis，返回 error 方便上层处理

func init() {
	err := InitDB()
	if err != nil {
		Error("InitDB error: %v", err)
	}
}

func InitDB() error {
	err := initMongo()
	if err != nil {
		return err
	}

	err = initRedis()
	if err != nil {
		return err
	}

	return nil
}

func initMongo() error {
	mongoCfg := Config.MongoDB
	clientOpts := options.Client().ApplyURI(mongoCfg.URI)
	if mongoCfg.TLS {
		clientOpts.SetTLSConfig(&tls.Config{})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		Error("MongoDB connect error: %v", err)
		return err
	}

	// 测试连接
	err = client.Ping(ctx, nil)
	if err != nil {
		Error("MongoDB ping error: %v", err)
		return err
	}

	MongoClient = client
	MongoDB = client.Database(mongoCfg.DB)
	Info("MongoDB connected: %s", mongoCfg.DB)
	return nil
}

func initRedis() error {
	redisCfg := Config.Redis

	options := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisCfg.Host, redisCfg.Port),
		DB:       redisCfg.DB,
		Password: redisCfg.Password,
	}

	if redisCfg.SSL {
		options.TLSConfig = &tls.Config{}
	}

	RedisClient = redis.NewClient(options)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		Error("Redis ping error: %v", err)
		return err
	}
	Info("Redis connected: %s", redisCfg.Host)
	return nil
}
