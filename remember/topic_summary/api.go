package topic_summary

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

type UploadRequest struct {
	SessionID string    `json:"session_id"`
	Messages  []Message `json:"messages"`
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

	r.Post("/topic_summary/upload", uploadHandler)                // 上传接口
	r.Get("/topic_summary/activate/{sessionID}", activateHandler) // 查询活跃话题接口
	r.Get("/topic_summary/search/{sessionID}", searchHandler)     // 搜索接口
	r.Delete("/topic_summary/delete/{sessionID}", deleteHandler)  // 删除接口
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

	if req.SessionID == "" || len(req.Messages) == 0 {
		resp := UploadResponse{
			Code: -1,
			Msg:  fmt.Sprintf("%s session_id and messages are required", SERVER_NAME),
			Data: struct{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// 构造 QueueMessage 对象
	msg := QueueMessage{
		TaskID:    GenerateUUID(),
		SessionID: req.SessionID,
		Messages:  req.Messages,
		Timestamp: time.Now().UTC().Unix(),
		Retry:     0,
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

// activateHandler 查询活跃话题信息
func activateHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		json.NewEncoder(w).Encode(QueryResponse{
			Code: -1,
			Msg:  "session_id is required",
			Data: struct{}{},
		})
		return
	}

	var topicInfo TopicInfo
	filter := map[string]string{"session_id": sessionID}
	err := DBClient.InfoCollection.FindOne(context.Background(), filter).Decode(&topicInfo)
	if err != nil {
		json.NewEncoder(w).Encode(QueryResponse{
			Code: -1,
			Msg:  "failed to get topic info: " + err.Error(),
			Data: struct{}{},
		})
		return
	}

	json.NewEncoder(w).Encode(QueryResponse{
		Code: 0,
		Msg:  "success",
		Data: topicInfo,
	})
}

// searchHandler 搜索话题接口（支持中英文查询）
func searchHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		json.NewEncoder(w).Encode(QueryResponse{
			Code: -1,
			Msg:  "session_id is required",
			Data: struct{}{},
		})
		return
	}

	query := r.URL.Query().Get("q")

	// 获取活跃话题列表
	var topicInfo TopicInfo
	filter := map[string]string{"session_id": sessionID}
	err := DBClient.InfoCollection.FindOne(context.Background(), filter).Decode(&topicInfo)
	var activeTopics []string
	if err == nil {
		for _, topic := range topicInfo.ActiveTopics {
			activeTopics = append(activeTopics, topic.Topic)
		}
	}

	// 暂不启用翻译功能，直接使用原查询
	// translatedQuery, err := TranslateQuery(query)
	// if err != nil {
	// 	// 翻译失败不影响搜索功能，使用原查询
	// 	Error("Translation failed for query '%s': %v", query, err)
	// 	translatedQuery = query
	// }
	// Info("Searching topics with query: original='%s', translated='%s'", query, translatedQuery)

	// 直接使用原查询进行搜索
	Info("Searching topics with query: '%s'", query)

	// 搜索话题
	results, err := DBClient.GetTopicSummary(context.Background(), sessionID, query, activeTopics)
	if err != nil {
		json.NewEncoder(w).Encode(QueryResponse{
			Code: -1,
			Msg:  "failed to search topics: " + err.Error(),
			Data: struct{}{},
		})
		return
	}

	json.NewEncoder(w).Encode(QueryResponse{
		Code: 0,
		Msg:  "success",
		Data: results,
	})
}

// deleteHandler 删除 所有会话归纳、会话info、队列中任务
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

	// 删除数据库记录（包括会话信息）
	if err := DBClient.DeleteSessionTopics(context.Background(), sessionID); err != nil {
		json.NewEncoder(w).Encode(QueryResponse{
			Code: -1,
			Msg:  "failed to delete session topics: " + err.Error(),
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
		Msg:  "session topics and queue messages deleted successfully",
		Data: struct{}{},
	})
}

// deleteTopicHandler 删除指定话题

/*
func deleteTopicHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	topic := chi.URLParam(r, "topic")
	if sessionID == "" || topic == "" {
		json.NewEncoder(w).Encode(QueryResponse{
			Code: -1,
			Msg:  "session_id and topic are required",
			Data: struct{}{},
		})
		return
	}

	// 删除指定话题
	if err := DBClient.DeleteTopic(context.Background(), sessionID, topic); err != nil {
		json.NewEncoder(w).Encode(QueryResponse{
			Code: -1,
			Msg:  "failed to delete topic: " + err.Error(),
			Data: struct{}{},
		})
		return
	}

	json.NewEncoder(w).Encode(QueryResponse{
		Code: 0,
		Msg:  "topic deleted successfully",
		Data: struct{}{},
	})
}
*/
