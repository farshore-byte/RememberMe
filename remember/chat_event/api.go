package chat_event

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// UploadRequest 上传接口请求体
type Message struct {
	Role    string `json:"role" bson:"role"`
	Content string `json:"content" bson:"content"`
}

type Conversation struct {
	Timestamp int64     `json:"timestamp" bson:"timestamp"` // 对话对的时间戳
	Messages  []Message `json:"messages" bson:"messages"`   // 一对对话（user + assistant）
}

type UploadRequest struct {
	SessionID    string        `json:"session_id"`
	Conversations []Conversation `json:"conversations"` // 一轮完整的对话（多个对话对）
}

// UploadResponse 上传接口响应（统一格式）
type UploadResponse struct {
	Code int         `json:"code"` // 0 成功, -1 失败
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// QueryResponse 查询接口响应
type QueryResponse struct {
	Code int         `json:"code"` // 0 成功, -1 失败
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// authMiddleware Bearer token鉴权中间件
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"code": -1, "msg": "Authorization header required"}`, http.StatusUnauthorized)
			return
		}

		// 检查Bearer token格式
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"code": -1, "msg": "Invalid authorization format"}`, http.StatusUnauthorized)
			return
		}

		// 验证token
		if parts[1] != Config.Auth.Token {
			http.Error(w, `{"code": -1, "msg": "Invalid token"}`, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RegisterRoutes 注册路由
func RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(authMiddleware) // 应用鉴权中间件

	r.Post("/chat_event/upload", uploadHandler)               // 上传接口
	r.Get("/chat_event/get/{sessionID}", queryHandler)        // 查询接口
	r.Delete("/chat_event/delete/{sessionID}", deleteHandler) // 删除接口
	return r
}

// uploadHandler 处理上传请求
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req UploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := UploadResponse{
			Code: -1,
			Msg:  "invalid request body",
			Data: struct{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	if req.SessionID == "" || len(req.Conversations) == 0 {
		resp := UploadResponse{
			Code: -1,
			Msg:  fmt.Sprintf("%s session_id and conversations are required", SERVER_NAME),
			Data: struct{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// 直接使用 conversations，保留对话对和时间戳信息
	msg := QueueMessage{
		TaskID:       GenerateUUID(),
		SessionID:    req.SessionID,
		Conversations: req.Conversations,
		Timestamp:    time.Now().UTC().Unix(),
		Retry:        0,
	}

	// 入队列
	if _, err := MessageQueue.Enqueue(ctx, msg); err != nil {
		resp := UploadResponse{
			Code: -1,
			Msg:  fmt.Sprintf("failed to %s enqueue", SERVER_NAME),
			Data: struct{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// 成功响应，返回 task_id
	resp := UploadResponse{
		Code: 0,
		Msg:  fmt.Sprintf("messages uploaded %s successfully", SERVER_NAME),
		Data: map[string]string{"task_id": msg.TaskID},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// queryHandler 查询用户画像
func queryHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		json.NewEncoder(w).Encode(QueryResponse{
			Code: -1,
			Msg:  "session_id is required",
			Data: struct{}{},
		})
		return
	}

	userPortrait, err := DBClient.GetSessionEvents(sessionID)
	if err != nil {
		json.NewEncoder(w).Encode(QueryResponse{
			Code: -1,
			Msg:  "failed to get user portrait: " + err.Error(),
			Data: struct{}{},
		})
		return
	}

	json.NewEncoder(w).Encode(QueryResponse{
		Code: 0,
		Msg:  "success",
		Data: userPortrait,
	})
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		json.NewEncoder(w).Encode(QueryResponse{
			Code: -1,
			Msg:  "session_id is required",
			Data: struct{}{},
		})
		return
	}

	// 删除数据库记录
	if err := DBClient.DeleteSessionEvents(sessionID); err != nil {
		json.NewEncoder(w).Encode(QueryResponse{
			Code: -1,
			Msg:  "failed to delete user portrait: " + err.Error(),
			Data: struct{}{},
		})
		return
	}

	// 删除队列中的消息
	ctx := context.Background()
	if err := MessageQueue.DeleteBySession(ctx, sessionID); err != nil {
		json.NewEncoder(w).Encode(QueryResponse{
			Code: -1,
			Msg:  "failed to delete messages from queue: " + err.Error(),
			Data: struct{}{},
		})
		return
	}

	json.NewEncoder(w).Encode(QueryResponse{
		Code: 0,
		Msg:  "user portrait and queue messages deleted successfully",
		Data: struct{}{},
	})
}
