package main

import (
	"fmt"
	"log"
	"net/http"
	"remember/openai"
)

func main() {
	// 初始化
	openai.InitLLM()
	// 注册路由
	router := openai.RegisterRoutes()
	// 初始化配置
	fmt.Println("Starting OpenAI Stream Completion Service...")
	fmt.Printf("Server URL: %s\n", openai.ServerURL)
	fmt.Printf("LLM Model: %s\n", openai.LLMModel)
	// 启动服务器
	port := openai.Config.Server.Openai
	addr := fmt.Sprintf(":%d", port)

	fmt.Printf("Server listening on port %d\n", port)
	fmt.Printf("API endpoint: http://localhost:%d/v1/response\n", port)

	log.Fatal(http.ListenAndServe(addr, router))
}
