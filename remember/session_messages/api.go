package session_messages

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
type UploadRequest struct {
	SessionID string                   `json:"session_id"`
	Messages  []map[string]interface{} `json:"messages"` // 支持 messages 格式
	TaskID    string                   `json:"task_id"`
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

	//------------------- 基本接口 ---------------------
	r.Post("/session_messages/upload", uploadHandler)               // 上传接口
	r.Get("/session_messages/get/{sessionID}", queryHandler)        // 查询接口
	r.Delete("/session_messages/delete/{sessionID}", deleteHandler) // 删除接口

	r.Get("/session_messages/count/{sessionID}", countHandler) // 查询当前会话中消息数量

	//---------------------  任务接口 ---------------------------
	r.Post("/session_messages/clean", cleanSsesionHandler) //  清理已处理的消息

	r.Post("/session_messages/mark_task", markEmptyTaskHandler) // 查找 taskN_id 为空并标记

	return r
}

// uploadHandler 处理上传请求
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	_, cancel := context.WithTimeout(r.Context(), 5*time.Second)
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

	if req.SessionID == "" {
		resp := UploadResponse{
			Code: -1,
			Msg:  fmt.Sprintf("%s session_id is required", SERVER_NAME),
			Data: struct{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// 处理 messages 数组，过滤非支持字段并确保成对
	processedMessages, err := processMessages(req.Messages)
	if err != nil {
		resp := UploadResponse{
			Code: -1,
			Msg:  fmt.Sprintf("invalid messages format: %s", err.Error()),
			Data: struct{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// 将处理后的消息保存到数据库
	var messageIDs []string
	for _, msg := range processedMessages {
		// task.... 默认为空
		message := MemoryMessage{
			ID:               GenerateUUID(),
			SessionID:        req.SessionID,
			UserContent:      msg.UserContent,
			AssistantContent: msg.AssistantContent,
			CreatedAt:        time.Now().UTC(),
			MessagesID:       req.TaskID,
			Status:           0, // 默认为待处理
		}

		if err := DBClient.InsertMessage(&message); err != nil {
			resp := UploadResponse{
				Code: -1,
				Msg:  fmt.Sprintf("failed to %s insert message", SERVER_NAME),
				Data: struct{}{},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		messageIDs = append(messageIDs, message.ID)
	}

	// 成功响应，返回 message_ids
	resp := UploadResponse{
		Code: 0,
		Msg:  fmt.Sprintf("messages uploaded %s successfully", SERVER_NAME),
		Data: map[string]interface{}{"message_ids": messageIDs, "count": len(messageIDs)},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// queryHandler 查询会话消息
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

	messages, err := DBClient.GetMessagesBySessionID(sessionID)
	if err != nil {
		json.NewEncoder(w).Encode(QueryResponse{
			Code: -1,
			Msg:  "failed to get messages: " + err.Error(),
			Data: struct{}{},
		})
		return
	}

	// 转换为 messages 格式: [{"role":"user","content":""},{"role":"assistant","content":""}]
	formattedMessages := formatMessagesToRoleContent(messages)

	json.NewEncoder(w).Encode(QueryResponse{
		Code: 0,
		Msg:  "success",
		Data: map[string]interface{}{
			"messages": formattedMessages,
		},
	})
}

// deleteHandler 删除指定 session_id 的所有消息
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
	if err := DBClient.DeleteMessagesBySessionID(sessionID); err != nil {
		json.NewEncoder(w).Encode(QueryResponse{
			Code: -1,
			Msg:  "failed to delete messages: " + err.Error(),
			Data: struct{}{},
		})
		return
	}

	json.NewEncoder(w).Encode(QueryResponse{
		Code: 0,
		Msg:  "messages deleted successfully",
		Data: struct{}{},
	})
}

func cleanSsesionHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error("clean session messages:invalid request body: " + err.Error())
		json.NewEncoder(w).Encode(QueryResponse{
			Code: -1,
			Msg:  "invalid request body: " + err.Error(),
			Data: struct{}{},
		})
		return
	}
	if req.SessionID == "" {
		Error("clean session messages: session_id is required")
		json.NewEncoder(w).Encode(QueryResponse{
			Code: -1,
			Msg:  "session_id is required",
			Data: struct{}{},
		})
		return
	}
	//清理
	sessionID := req.SessionID
	// 清理数据库记录
	if err := DBClient.clearSessionMessages(sessionID); err != nil {
		json.NewEncoder(w).Encode(QueryResponse{
			Code: -1,
			Msg:  "failed to clean messages: " + err.Error(),
			Data: struct{}{},
		})
		return
	}
	// 成功响应
	json.NewEncoder(w).Encode(QueryResponse{
		Code: 0,
		Msg:  fmt.Sprintf("%s messages cleaned successfully", SERVER_NAME),
		Data: struct{}{},
	})
}

func countHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		json.NewEncoder(w).Encode(QueryResponse{
			Code: -1,
			Msg:  "session_id is required",
			Data: struct{}{},
		})
		return
	}
	count, err := DBClient.CountMessagesBySessionID(sessionID)
	if err != nil {
		json.NewEncoder(w).Encode(QueryResponse{
			Code: -1,
			Msg:  "failed to count messages: " + err.Error(),
			Data: struct{}{},
		})
		return
	}
	json.NewEncoder(w).Encode(QueryResponse{
		Code: 0,
		Msg:  "success",
		Data: map[string]interface{}{
			"count": count,
		},
	})
}

// MarkTaskRequest 请求体
type MarkTaskRequest struct {
	SessionID string `json:"session_id"`
	TaskIndex int    `json:"task_index"` // 1~4
	TaskID    string `json:"task_id"`
}

// markEmptyTaskHandler 查找指定 session 下 taskN_id 为空的消息并标记为 taskID
func markEmptyTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req MarkTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(QueryResponse{
			Code: -1,
			Msg:  "invalid request body: " + err.Error(),
			Data: struct{}{},
		})
		return
	}

	if req.SessionID == "" || req.TaskIndex < 1 || req.TaskIndex > 4 || req.TaskID == "" {
		json.NewEncoder(w).Encode(QueryResponse{
			Code: -1,
			Msg:  "invalid params: session_id, task_index (1~4), task_id required",
			Data: struct{}{},
		})
		return
	}

	// 调用 DBClient 的新方法
	messages, err := DBClient.FindAndMarkMessagesWithoutTaskID(req.SessionID, req.TaskIndex, req.TaskID)
	// 转换为 messages 格式: [{"role":"user","content":"","timestamp":""},{"role":"assistant","content":"","timestamp}]
	formattedMessages := formatMessagesToRoleContent(messages)
	if err != nil {
		json.NewEncoder(w).Encode(QueryResponse{
			Code: -1,
			Msg:  "failed to mark messages: " + err.Error(),
			Data: struct{}{},
		})
		return
	}

	json.NewEncoder(w).Encode(QueryResponse{
		Code: 0,
		Msg:  fmt.Sprintf("successfully marked %d messages for task%d", len(messages), req.TaskIndex),
		Data: map[string]interface{}{
			"messages": formattedMessages,
		},
	})
}
