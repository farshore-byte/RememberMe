package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"remember/config"
	"remember/session_messages"
	"syscall"
	"time"
)

func main() {

	// 注册 HTTP 路由
	r := session_messages.RegisterRoutes()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Config.Server.SessionMessages), // 监听端口,,
		Handler: r,
	}

	// 启动 HTTP 服务
	go func() {
		log.Printf("✅ Session Messages API running at http://localhost:%d", config.Config.Server.SessionMessages)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	// 捕获系统信号，用于优雅退出
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("🛑 Signal received, shutting down...")

	// 优雅关闭 HTTP 服务
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server Shutdown: %v", err)
	}
	log.Println("✅ HTTP server stopped gracefully")
}
