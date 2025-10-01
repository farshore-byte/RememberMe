package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/spf13/viper"
)

// 全局变量
var (
	LLMModel     string
	OpenAIClient openai.Client
	ServerURL    string
)

// InitLLM 初始化 OpenAI Client
func InitLLM() {
	// 确保配置已经加载
	if Config.Server.Main == 0 {
		// 如果配置未加载，尝试重新加载
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("..")
		viper.AddConfigPath("../..")
		if err := viper.ReadInConfig(); err == nil {
			viper.Unmarshal(&Config)
		}
	}

	OpenAIClient = openai.NewClient(
		option.WithAPIKey(Config.LLM.APIKey),
		option.WithBaseURL(Config.LLM.BaseURL),
	)
	LLMModel = Config.LLM.ModelID
	ServerURL = fmt.Sprintf("http://localhost:%d", Config.Server.Main)
}

// 请求结构
type StreamCompletionRequest struct {
	Query        string `json:"query"`
	SessionID    string `json:"session_id,omitempty"`
	UserID       string `json:"user_id,omitempty"`
	RoleID       string `json:"role_id,omitempty"`
	GroupID      string `json:"group_id,omitempty"`
	RolePrompt   string `json:"role_prompt,omitempty"`
	FirstMessage string `json:"first_message,omitempty"`
	Stream       *bool  `json:"stream,omitempty"`
}

// 响应结构
type StreamCompletionResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Content string `json:"content,omitempty"`
	} `json:"data,omitempty"`
}

// 消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// apply_memory 响应结构
type ApplyMemoryResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		SystemPrompt string    `json:"system_prompt"`
		Messages     []Message `json:"messages"`
	} `json:"data"`
}

// 上传请求结构
type UploadRequest struct {
	SessionID string    `json:"session_id"`
	UserID    string    `json:"user_id"`
	RoleID    string    `json:"role_id"`
	GroupID   string    `json:"group_id"`
	Messages  []Message `json:"messages"`
}

// 上传响应结构
type UploadResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// 注册路由
func RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(authMiddleware)

	// 流式完成接口
	r.Post("/v1/response", streamCompletionHandler)

	return r
}

// 鉴权中间件
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"code": -1, "msg": "Authorization header required"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"code": -1, "msg": "Invalid authorization format"}`, http.StatusUnauthorized)
			return
		}

		if parts[1] != Config.Auth.Token {
			http.Error(w, `{"code": -1, "msg": "Invalid token"}`, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// 流式完成处理函数
func streamCompletionHandler(w http.ResponseWriter, r *http.Request) {
	var req StreamCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, StreamCompletionResponse{
			Code: -1,
			Msg:  "参数解析错误: " + err.Error(),
		})
		return
	}

	if req.Query == "" {
		writeJSON(w, StreamCompletionResponse{
			Code: -1,
			Msg:  "query is required",
		})
		return
	}

	// 1. 调用server的apply_memory接口获取系统提示词和消息
	systemPrompt, messages, err := getSystemPromptAndMessages(req)
	if err != nil {
		writeJSON(w, StreamCompletionResponse{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	// 检查是否启用流式
	useStream := req.Stream != nil && *req.Stream

	if useStream {
		// 流式模式
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		// 调用流式生成函数
		err = generateStreamResponse(r.Context(), systemPrompt, messages, req.Query, w, flusher, req)
		if err != nil {
			sendErrorEvent(w, flusher, err.Error())
			return
		}

		// 流式模式下，上传对话在generateStreamResponse中处理
	} else {
		// 非流式模式
		responseContent, err := generateNonStreamResponse(r.Context(), systemPrompt, messages, req.Query)
		if err != nil {
			writeJSON(w, StreamCompletionResponse{
				Code: -1,
				Msg:  err.Error(),
			})
			return
		}

		// 返回完整响应
		resp := StreamCompletionResponse{
			Code: 0,
			Msg:  "success",
		}
		resp.Data.Content = responseContent
		writeJSON(w, resp)

		// 上传对话
		go uploadConversation(req, messages, responseContent)
	}
}

// 获取系统提示词和消息
func getSystemPromptAndMessages(req StreamCompletionRequest) (string, []Message, error) {
	applyReq := map[string]interface{}{
		"query":       req.Query,
		"session_id":  req.SessionID,
		"user_id":     req.UserID,
		"role_id":     req.RoleID,
		"group_id":    req.GroupID,
		"role_prompt": req.RolePrompt,
	}

	applyResp, err := callServerAPI("/memory/apply", applyReq)
	if err != nil {
		return "", nil, fmt.Errorf("调用apply_memory接口失败: %w", err)
	}

	var applyData ApplyMemoryResponse
	if err := json.Unmarshal(applyResp, &applyData); err != nil {
		return "", nil, fmt.Errorf("解析apply_memory响应失败: %w", err)
	}

	if applyData.Code != 0 {
		return "", nil, fmt.Errorf("apply_memory接口返回错误: %s", applyData.Msg)
	}

	// 如果apply_memory返回的messages为空，且提供了first_message，则上传初始对话到server
	if len(applyData.Data.Messages) == 0 && req.FirstMessage != "" {
		// 创建初始对话：空用户消息 + first_message作为助手回复
		Warn("apply接口没有消息返回，默认为首次对话，创建首轮对话：空用户消息+first message!!!")
		initialMessages := []Message{
			{
				Role:    "user",
				Content: "",
			},
			{
				Role:    "assistant",
				Content: req.FirstMessage,
			},
		}

		// 上传初始对话到server（但不作为回复返回）
		Info("first message upload to server")
		go uploadInitialConversation(req, initialMessages)

		// 返回空的messages，让OpenAI生成新的回复
		return applyData.Data.SystemPrompt, []Message{}, nil
	}
	Info(fmt.Sprintf("%s generate system prompt: %s", SERVER_NAME, applyData.Data.SystemPrompt))

	return applyData.Data.SystemPrompt, applyData.Data.Messages, nil
}

// 生成非流式响应
func generateNonStreamResponse(ctx context.Context, systemPrompt string, messages []Message, query string) (string, error) {
	// 构建消息列表
	chatMessages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
	}

	// 添加历史消息
	for _, msg := range messages {
		switch msg.Role {
		case "user":
			chatMessages = append(chatMessages, openai.UserMessage(msg.Content))
		case "assistant":
			chatMessages = append(chatMessages, openai.AssistantMessage(msg.Content))
		case "system":
			chatMessages = append(chatMessages, openai.SystemMessage(msg.Content))
		}
	}

	// 添加当前查询
	chatMessages = append(chatMessages, openai.UserMessage(query))

	// 调用OpenAI非流式接口
	resp, err := OpenAIClient.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Messages: chatMessages,
			Model:    LLMModel,
		},
	)
	if err != nil {
		return "", fmt.Errorf("OpenAI调用失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response from model")
	}

	return resp.Choices[0].Message.Content, nil
}

// 生成流式响应
func generateStreamResponse(ctx context.Context, systemPrompt string, messages []Message, query string, w http.ResponseWriter, flusher http.Flusher, req StreamCompletionRequest) error {
	// 构建消息列表
	chatMessages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
	}

	// 添加历史消息
	for _, msg := range messages {
		switch msg.Role {
		case "user":
			chatMessages = append(chatMessages, openai.UserMessage(msg.Content))
		case "assistant":
			chatMessages = append(chatMessages, openai.AssistantMessage(msg.Content))
		case "system":
			chatMessages = append(chatMessages, openai.SystemMessage(msg.Content))
		}
	}

	// 添加当前查询
	chatMessages = append(chatMessages, openai.UserMessage(query))

	// 调用OpenAI流式接口
	stream := OpenAIClient.Chat.Completions.NewStreaming(
		ctx,
		openai.ChatCompletionNewParams{
			Messages: chatMessages,
			Model:    LLMModel,
		},
	)

	// 使用accumulator来收集流式响应
	acc := openai.ChatCompletionAccumulator{}
	var fullResponse strings.Builder

	// 流式返回响应内容
	for stream.Next() {
		chunk := stream.Current()
		acc.AddChunk(chunk)

		// 发送流式数据
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			content := chunk.Choices[0].Delta.Content
			if content != "" {
				// 累积完整回复
				fullResponse.WriteString(content)

				streamResp := StreamCompletionResponse{
					Code: 0,
					Msg:  "success",
				}
				streamResp.Data.Content = content

				streamJSON, _ := json.Marshal(streamResp)
				fmt.Fprintf(w, "data: %s\n\n", streamJSON)
				flusher.Flush()
			}
		}
	}

	if err := stream.Err(); err != nil {
		return fmt.Errorf("流式响应错误: %w", err)
	}

	// 发送结束事件
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()

	// 对话结束后，上传当轮用户问题和模型回复
	if fullResponse.Len() > 0 {
		responseContent := fullResponse.String()
		fmt.Printf("对话结束，准备上传对话记录:\n")
		fmt.Printf("Query: %s\n", req.Query)
		fmt.Printf("Response: %s\n", responseContent)
		fmt.Printf("SessionID: %s, UserID: %s, RoleID: %s\n", req.SessionID, req.UserID, req.RoleID)

		go uploadCurrentConversation(req, query, responseContent)
	}

	return nil
}

// 发送错误事件
func sendErrorEvent(w http.ResponseWriter, flusher http.Flusher, errorMsg string) {
	errorResp := StreamCompletionResponse{
		Code: -1,
		Msg:  errorMsg,
	}

	errorJSON, _ := json.Marshal(errorResp)
	fmt.Fprintf(w, "data: %s\n\n", errorJSON)
	flusher.Flush()
}

// 发送流式响应
func sendStreamResponse(w http.ResponseWriter, flusher http.Flusher, content string) {
	// 发送内容事件
	contentResp := StreamCompletionResponse{
		Code: 0,
		Msg:  "success",
	}
	contentResp.Data.Content = content

	contentJSON, _ := json.Marshal(contentResp)
	fmt.Fprintf(w, "data: %s\n\n", contentJSON)
	flusher.Flush()

	// 发送结束事件
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

// 上传对话到server（上传完整历史）
func uploadConversation(req StreamCompletionRequest, historyMessages []Message, responseContent string) {
	// 构建新的消息列表
	newMessages := append(historyMessages, Message{
		Role:    "user",
		Content: req.Query,
	}, Message{
		Role:    "assistant",
		Content: responseContent,
	})

	uploadReq := UploadRequest{
		SessionID: req.SessionID,
		UserID:    req.UserID,
		RoleID:    req.RoleID,
		GroupID:   req.GroupID,
		Messages:  newMessages,
	}

	_, err := callServerAPI("/memory/upload", uploadReq)
	if err != nil {
		fmt.Printf("上传对话失败: %v\n", err)
	}
}

// 上传初始对话到server（使用first_message创建初始对话）
func uploadInitialConversation(req StreamCompletionRequest, initialMessages []Message) {
	uploadReq := UploadRequest{
		SessionID: req.SessionID,
		UserID:    req.UserID,
		RoleID:    req.RoleID,
		GroupID:   req.GroupID,
		Messages:  initialMessages,
	}

	_, err := callServerAPI("/memory/upload", uploadReq)
	if err != nil {
		fmt.Printf("上传初始对话失败: %v\n", err)
	} else {
		fmt.Printf("初始对话上传成功: 使用first_message创建初始对话\n")
	}
}

// 上传当轮对话到server（只上传当前query和response）
func uploadCurrentConversation(req StreamCompletionRequest, query string, responseContent string) {
	// 只构建当前轮次的消息
	newMessages := []Message{
		{
			Role:    "user",
			Content: query,
		},
		{
			Role:    "assistant",
			Content: responseContent,
		},
	}

	uploadReq := UploadRequest{
		SessionID: req.SessionID,
		UserID:    req.UserID,
		RoleID:    req.RoleID,
		GroupID:   req.GroupID,
		Messages:  newMessages,
	}

	_, err := callServerAPI("/memory/upload", uploadReq)
	if err != nil {
		fmt.Printf("上传当轮对话失败: %v\n", err)
	} else {
		fmt.Printf("当轮对话上传成功\n")
	}
}

// 调用server API
func callServerAPI(endpoint string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	url := ServerURL + endpoint
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+Config.Auth.Token)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// 写入JSON响应
func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
