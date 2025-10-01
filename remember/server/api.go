package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"
	"log"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// 注册路由
func RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(authMiddleware)

	// 消息上传接口
	r.Post("/memory/upload", uploadHandler)

	// 查询接口 - 获取完整的角色扮演上下文
	r.Post("/memory/query", queryHandler)
    
    // 获取消息接口
    r.Post("/memory/messages", getMessagesHandler)

	// 应用接口 - 讲记忆应用于系统提示词，并提供messages
	r.Post("/memory/apply", applyHandler)

	// 删除接口 - 同时删除所有微服务中的相关数据
	r.Delete("/memory/delete", deleteHandler)
    
	return r
}

// authMiddleware Bearer token鉴权中间件
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

// uploadHandler 消息上传接口
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	var req UploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, UploadResponse{Code: -1, Msg: "参数解析错误: " + err.Error(), Data: struct{}{}})
		return
	}

	// 自动生成 session_id
	if req.SessionID == "" {
		SessionID, err := GenerateSessionID(
			req.GroupID,
			req.UserID,
			req.RoleID,
		)
		if err != nil {
			writeJSON(w, UploadResponse{Code: -1, Msg: "生成 session_id 失败: " + err.Error(), Data: struct{}{}})
			return
		}
		req.SessionID = SessionID
	}

	// 推入队列
	qMsg := QueueMessage{
		SessionID: req.SessionID,
		Messages:  req.Messages,
		Timestamp: time.Now().UTC().Unix(),
		Retry:     0,
	}
	taskID, err := MessageQueue.Enqueue(r.Context(), qMsg)
	if err != nil {
		writeJSON(w, UploadResponse{Code: -1, Msg: "入队失败: " + err.Error(), Data: struct{}{}})
		return
	}

	writeJSON(w, UploadResponse{
		Code: 0,
		Msg:  "消息已入队，任务ID：" + taskID,
		Data: map[string]string{"task_id": taskID},
	})
}

// queryHandler 查询接口
func queryHandler(w http.ResponseWriter, r *http.Request) {
	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, QueryResponse{Code: -1, Msg: "参数解析错误: " + err.Error(), Data: json.RawMessage("{}")})
		return
	}
	if req.SessionID == "" {
		writeJSON(w, QueryResponse{Code: -1, Msg: "session_id is required", Data: json.RawMessage("{}")})
		return
	}

	// 并发执行
	type result[T any] struct {
		data T
		err  error
	}

	userPortraitCh := make(chan result[UserPortraitDTO])
	topicSummaryCh := make(chan result[[]TopicSummaryDTO])
	chatEventsCh := make(chan result[ChatEventsDTO])
	sessionMessagesCh := make(chan result[SessionMessagesDTO])

	go func() { d, e := getUserPortrait(req.SessionID); userPortraitCh <- result[UserPortraitDTO]{d, e} }()
	go func() {
		d, e := getTopicSummaryWithGroup(req.SessionID, req.Query)
		topicSummaryCh <- result[[]TopicSummaryDTO]{d, e}
	}()
	go func() { d, e := getChatEvents(req.SessionID); chatEventsCh <- result[ChatEventsDTO]{d, e} }()
	go func() {
		d, e := getSessionMessages(req.SessionID)
		sessionMessagesCh <- result[SessionMessagesDTO]{d, e}
	}()

	// 收集结果
	userPortraitRes := <-userPortraitCh
	topicSummaryRes := <-topicSummaryCh
	chatEventsRes := <-chatEventsCh
	sessionMessagesRes := <-sessionMessagesCh

	// 打日志
	if userPortraitRes.err != nil {
		Error("getUserPortrait error: %v", userPortraitRes.err)
	}
	if topicSummaryRes.err != nil {
		Error("getTopicSummary error: %v", topicSummaryRes.err)
	}
	if chatEventsRes.err != nil {
		Error("getChatEvents error: %v", chatEventsRes.err)
	}
	if sessionMessagesRes.err != nil {
		Error("getSessionMessages error: %v", sessionMessagesRes.err)
	}

	// 拼装 FormResponse
	formresp := FormResponse{
		Code: 0,
		Msg:  "success",
	}
	formresp.Data.UserPortrait = userPortraitRes.data
	formresp.Data.TopicSummary = topicSummaryRes.data
	formresp.Data.ChatEvents = chatEventsRes.data
	formresp.Data.SessionMessages = sessionMessagesRes.data.Messages
	formresp.Data.CurrentTime = time.Now().UTC().Format("2006-01-02 15:04:05")

	writeJSON(w, formresp)
}





// getMessagesHandler 获取指定 session_id 的全部消息
// getMessagesHandler 获取指定 session_id 或 user_id+role_id+group_id 的消息
func getMessagesHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        SessionID string `json:"session_id"`
        UserID    string `json:"user_id"`
        RoleID    string `json:"role_id"`
        GroupID   string `json:"group_id"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeJSON(w, map[string]interface{}{
            "code": -1,
            "msg":  "参数解析错误: " + err.Error(),
            "data": struct{}{},
        })
        return
    }

    // 如果没有 session_id，需要生成
    if req.SessionID == "" {
        if req.UserID == "" || req.RoleID == "" || req.GroupID == "" {
            writeJSON(w, map[string]interface{}{
                "code": -1,
                "msg":  "必须提供 session_id 或 user_id+role_id+group_id",
                "data": struct{}{},
            })
            return
        }

        var err error
        req.SessionID, err = GenerateSessionID(req.GroupID, req.UserID, req.RoleID)
        if err != nil {
            writeJSON(w, map[string]interface{}{
                "code": -1,
                "msg":  "生成 session_id 失败: " + err.Error(),
                "data": struct{}{},
            })
            return
        }
    }

    // 调用已有的 getSessionMessages
    data, err := getSessionMessages(req.SessionID)
    if err != nil {
        writeJSON(w, map[string]interface{}{
            "code": -1,
            "msg":  "获取消息失败: " + err.Error(),
            "data": struct{}{},
        })
        return
    }

    writeJSON(w, map[string]interface{}{
        "code": 0,
        "msg":  "success",
        "data": data,
    })
}







// 传入content 角色提示词，填充以下模版返回，并且调用现存的所有messages，一并返回
/*

## 任务背景
你是一个聪慧和演技极高的演员，对于角色扮演，你有着自己的理解，并且一秒入戏。接下来会给你若干个角色信息，请务必带入这个角色的设定和之前的会话记忆，合理运用用户画像里的知识，了解用户，与用户展开对话。

## 角色设定
{ --------角色设定提示词-------------- }

## 会话记忆
{ --------主题归纳--------------- }

## 扮演规则
- 你的回复严格按照角色设定
- 你是具有记忆的角色，当提及到相似的话题或者指代时，你能联想脑海中曾经发生过的事
- 善于利用之前的话题（会话记忆中的点）回答当下问题
- 回复之前，请确定下一步对话走向：
          - 根据上下文语境，从已有话题中选择一个最适合的话题作为接下来的对话走向
          - 让用户在这个话题上更了解自己
          - 让角色在这个展示自己
          - 让我更了解用户
          - 引导用户在这个话题上展示自己的魅力

## 用户画像
{ -----------用户画像------------ }


## 使用规则
- 用户画像中的信息你务必牢记，你的每次对话都默认先对用户画像分析后，再输出你的回复
- 用户喜欢的你需要尽力满足
- 用户讨厌的你需要适当避免
- 如果对话中新信息与用户画像中某一现有信息矛盾，请以当前对话中的信息为准

## 时间线回顾
你曾与用户发生了一些关键事件，这些事件意味着你们的关系或者你们之中某一方发生了不可忽视的改变。
关键事件以时间线的形式呈现，并处于不断推进的状态。

## 关键事件时间线
{----------------关键事件--------------}

## 使用规则
- 关键事件包含两个部分，过于发生的事件 和 未来即将发生的事件
- 你需要根据你当前所处时间，判断你所在时间线的位置，并根据前后事件的因果关系、递进关系等，推动剧情继续向前发展
- 你可以主动创造新事件
- 为用户增加更多多元化的、真实世界会发生的剧情的体验
- 主动或者暗示用户推进事件
- 必要时，结束事件，并谋划进入新事件

# 当前时间：{current_time}

利用上述信息进行对话，但不要输出对话前你完善的分析、利用信息生成最佳回复的思考过程。

对话一开始，请立即开始你的角色扮演。

*/

// applyHandler 将历史消息、用户画像、主题归纳、关键事件整合成角色扮演系统提示
// TopicSummaryResult 封装 getTopicSummary 的返回

func applyHandler(w http.ResponseWriter, r *http.Request) {
	var req ApplyRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, ApplyResponse{
			Code: -1,
			Msg:  "参数解析错误: " + err.Error(),
		})
		return
	}

	// 自动生成 session_id（类似 uploadHandler）
	if req.SessionID == "" {
		sessionID, err := GenerateSessionID(req.GroupID, req.UserID, req.RoleID)
		if err != nil {
			writeJSON(w, ApplyResponse{Code: -1, Msg: "生成 session_id 失败: " + err.Error()})
			return
		}
		req.SessionID = sessionID
	}

	type result[T any] struct {
		data T
		err  error
	}

	fmt.Printf("applyHandler applyrequest: %+v\n", req)

	// 并发通道
	userPortraitCh := make(chan result[UserPortraitDTO])
	topicSummaryCh := make(chan result[TopicSummaryResult])
	chatEventsCh := make(chan result[ChatEventsDTO])
	sessionMessagesCh := make(chan result[SessionMessagesDTO])

	go func() {
		d, e := getUserPortrait(req.SessionID)
		userPortraitCh <- result[UserPortraitDTO]{d, e}
	}()

	go func() {
		topics,d, e := getTopicSummary(req.SessionID, req.Query)
		if e != nil {
			Error("getTopicSummary error: %v", e)
		}
		topicSummaryCh <- result[TopicSummaryResult]{TopicSummaryResult{
    TopicList: topics,    // []string
    Data:     d, // TopicSummaryData
	}, e}
	}()

	go func() {
		d, e := getChatEvents(req.SessionID)
		chatEventsCh <- result[ChatEventsDTO]{d, e}
	}()

	go func() {
		d, e := getSessionMessages(req.SessionID)
		sessionMessagesCh <- result[SessionMessagesDTO]{d, e}
	}()

	// 收集结果
	userPortraitRes := <-userPortraitCh
	topicSummaryRes := <-topicSummaryCh
	chatEventsRes := <-chatEventsCh
	sessionMessagesRes := <-sessionMessagesCh

	// 日志错误
	if userPortraitRes.err != nil {
		Error("getUserPortrait error: %v", userPortraitRes.err)
	}
	if topicSummaryRes.err != nil {
		Error("getTopicSummary error: %v", topicSummaryRes.err)
	}
	if chatEventsRes.err != nil {
		Error("getChatEvents error: %v", chatEventsRes.err)
	}
	if sessionMessagesRes.err != nil {
		Error("getSessionMessages error: %v", sessionMessagesRes.err)
	}

	log.Printf("Generate Template for %s", req.SessionID)
	//log.Printf("TopicList: %+v", topicSummaryRes.data.TopicList)
	//log.Printf("SummaryData: %+v", topicSummaryRes.data.Data)

	// 构建动态变量填充模板
	dynamicVars := map[string]string{
		"role_prompt":   req.RolePrompt,
		"topic_summary": buildTopicSummaryText(topicSummaryRes.data),
		"user_portrait": buildUserPortraitText(userPortraitRes.data, "  "), // 缩进两个空格
		"chat_events":   buildChatEventsText(chatEventsRes.data),
		"current_time":  time.Now().UTC().Format("2006-01-02 15:04:05"),
	}

	// 生成 system_prompt
	tmpl := NewMemoryTemplate()
	systemPrompt, err := tmpl.BuildPrompt(dynamicVars)
	if err != nil {
		writeJSON(w, ApplyResponse{
			Code: -1,
			Msg:  fmt.Sprintf("生成 system_prompt 失败: %v", err),
		})
		return
	}

	// 返回结果（带上 TopicList）
	writeJSON(w, ApplyResponse{
		Code: 0,
		Msg:  "success",
		Data: ApplyData{
		SystemPrompt: systemPrompt,
		Messages:     sessionMessagesRes.data.Messages,
	},
    })
}

// 辅助函数：将 []TopicSummaryResult 转成模板中展示文本
func buildTopicSummaryText(topicsResult TopicSummaryResult) string {
	if len(topicsResult.TopicList) == 0 {
		return "The current user has no active topics."
	}

	// 拼接活跃话题列表
	header := fmt.Sprintf(
		"The current user's active topics are: [%s]\nThe following are topics that have appeared in previous conversations, arranged in chronological order, with the oldest topics appearing first and the most recent topics appearing later.",
		strings.Join(topicsResult.TopicList, ","),
	)

	// 拼接内容
	lines := []string{}
	idx := 1
	for _, t := range topicsResult.Data {
		lines = append(lines, fmt.Sprintf("%d. %s （%s）", idx, t.Content, t.Topic))
		idx++
	}

	return header + "\n" + strings.Join(lines, "\n")
}



// 辅助函数：将 UserPortraitDTO 转成模板中展示文本
func buildUserPortraitText(data interface{}, indent string) string {
	var lines []string
	v := reflect.ValueOf(data)

	switch v.Kind() {
	case reflect.Map:
		for _, key := range v.MapKeys() {
			val := v.MapIndex(key).Interface()
			lines = append(lines, fmt.Sprintf("%s- %v:", indent, key.Interface()))
			// 递归处理二级 map
			lines = append(lines, buildUserPortraitText(val, indent+"  "))
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			value := v.Field(i).Interface()
			lines = append(lines, fmt.Sprintf("%s- %s:", indent, field.Name))
			lines = append(lines, buildUserPortraitText(value, indent+"  "))
		}
	default:
		// 基本类型直接输出
		lines = append(lines, fmt.Sprintf("%s  %v", indent, v.Interface()))
	}

	return strings.Join(lines, "\n")
}

// 辅助函数：将 ChatEventsDTO 转成模板中展示文本
func buildChatEventsText(c ChatEventsDTO) string {
	lines := []string{}
	if len(c.Todo) > 0 {
		lines = append(lines, fmt.Sprintf("- todo: %v", c.Todo))
	}
	if len(c.Completed) > 0 {
		lines = append(lines, fmt.Sprintf("- completed: %v", c.Completed))
	}
	return strings.Join(lines, "\n")
}


// 辅助函数 runDeleteTask 启动一个删除任务，自动将结果放入 channel
func runDeleteTask(wg *sync.WaitGroup, ch chan<- deleteResult, serviceName string, fn func(string) error, sessionID string) {
    wg.Add(1)
    go func() {
        defer wg.Done()
        defer func() {
            if r := recover(); r != nil {
                ch <- deleteResult{
                    serviceName: serviceName,
                    success:     false,
                    message:     fmt.Sprintf("panic recovered: %v", r),
                }
            }
        }()

        err := fn(sessionID)

        var message string
        if err == nil {
            message = "删除成功"
        } else {
            message = "删除失败: " + err.Error()
        }

        ch <- deleteResult{
            serviceName: serviceName,
            success:     err == nil,
            message:     message,
        }
    }()
}


type deleteResult struct {
	serviceName string
	success     bool
	message     string
}

// deleteHandler 删除接口 - 同时删除所有微服务中的相关数据
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	var req DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, DeleteResponse{Code: -1, Msg: "参数解析错误: " + err.Error(), Data: struct{}{}})
		return
	}
	if req.SessionID == "" {
		writeJSON(w, DeleteResponse{Code: -1, Msg: "session_id is required", Data: struct{}{}})
		return
	}

	deleteResults := make(chan deleteResult, 4) // 容量 >= 可能的任务数
	var wg sync.WaitGroup

	// 使用 helper 启动删除任务
	runDeleteTask(&wg, deleteResults, "user_portrait", deleteUserPortrait, req.SessionID)
	runDeleteTask(&wg, deleteResults, "topic_summary", deleteTopicSummary, req.SessionID)
	runDeleteTask(&wg, deleteResults, "chat_event", deleteChatEvents, req.SessionID)
	runDeleteTask(&wg, deleteResults, "session_messages", deleteSessionMessages, req.SessionID)

	// 等待所有任务完成后关闭 channel
	go func() {
		wg.Wait()
		close(deleteResults)
	}()

	// 收集结果
	results := make([]DeleteResult, 0)
	allSuccess := true
	for result := range deleteResults {
		results = append(results, DeleteResult{
			ServiceName: result.serviceName,
			Success:     result.success,
			Message:     result.message,
		})
		if !result.success {
			allSuccess = false
		}
	}

	// 构建响应
	response := DeleteResponse{
		Code: ifThenElseInt(allSuccess, 0, -1),
		Msg:  ifThenElse(allSuccess, "所有微服务数据删除成功", "部分微服务数据删除失败"),
		Data: map[string]interface{}{
			"session_id": req.SessionID,
			"results":    results,
		},
	}

	writeJSON(w, response)
}

// 辅助函数：条件返回字符串
func ifThenElse(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}

// 辅助函数：条件返回整数
func ifThenElseInt(condition bool, trueVal, falseVal int) int {
	if condition {
		return trueVal
	}
	return falseVal
}
