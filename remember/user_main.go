package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"remember/config"
	"remember/user_poritrait"
	"syscall"
	"time"
)

func main() {
	numWorkers := 100  // 你想启动的 Worker 数量
	workers := make([]*user_poritrait.Worker, 0, numWorkers)

	// 启动多个 Worker
	for i := 0; i < numWorkers; i++ {
		w := user_poritrait.NewWorker(1 * time.Second)
		w.Start()
		workers = append(workers, w)
		log.Printf("✅ Worker %d started\n", i+1)
	}
	// 启动队列监控
	monitor := &user_poritrait.QueueMonitor{
		Queue:    user_poritrait.MessageQueue,
		MaxLen:   user_poritrait.Queue_MAXLEN,                   // 队列长度阈值
		Interval: user_poritrait.Monitor_Interval * time.Second, // 检查间隔
	}
	monitor.Start()
	log.Println("✅ Queue monitor started")
	// 注册 HTTP 路由
	r := user_poritrait.RegisterRoutes()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Config.Server.UserPortrait), // 监听端口,//Addr:    ":7004",
		Handler: r,
	}

	// 启动 HTTP 服务
	go func() {
		//log.Println("✅ User Portrait API running at http://localhost:7004")
		log.Printf("✅ User Portrait API running at http://localhost:%d", config.Config.Server.UserPortrait)
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
