package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"remember/config"
	"remember/topic_summary"
	"syscall"
	"time"
)

func main() {
	numWorkers := 100   // 你想启动的 Worker 数量
	workers := make([]*topic_summary.Worker, 0, numWorkers)

	// 启动多个 Worker
	for i := 0; i < numWorkers; i++ {
		w := topic_summary.NewWorker(1 * time.Second)
		w.Start()
		workers = append(workers, w)
		log.Printf("✅ Worker %d started\n", i+1)
	}

	// 启动队列监控
	monitor := &topic_summary.QueueMonitor{
		Queue:    topic_summary.MessageQueue,
		MaxLen:   topic_summary.Queue_MAXLEN,                   // 队列长度阈值
		Interval: topic_summary.Monitor_Interval * time.Second, // 检查间隔
	}
	monitor.Start()
	log.Println("✅ Queue monitor started")

	// 注册 HTTP 路由
	r := topic_summary.RegisterRoutes()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Config.Server.TopicSummary), // 监听端口,//Addr:    ":7006", // 使用不同的端口
		Handler: r,
	}

	// 启动 HTTP 服务
	go func() {
		log.Printf("✅ Topic Summary API running at http://localhost:%d", config.Config.Server.TopicSummary)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	// 捕获系统信号，用于优雅退出
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("🛑 Signal received, shutting down...")

	// 停止所有 Worker
	for i, w := range workers {
		w.Stop()
		log.Printf("✅ Worker %d stopped gracefully\n", i+1)
	}

	// 优雅关闭 HTTP 服务
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server Shutdown: %v", err)
	}
	log.Println("✅ HTTP server stopped gracefully")
}
