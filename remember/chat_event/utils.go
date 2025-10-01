package chat_event

// 通用组装函数，支持变量插入不同的系统提示词模版，返回组装过后的系统提示词
import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"regexp"
	"strings"
	"time"
)

/*

两个函数拼装逻辑：
1. systempromptcompose 所有占位符都要匹配到，如果缺少变量，报错
2. systempromptcomposestatic 缺少变量跳过，不报错


*/

// 通用函数组装
func SystemPromptCompose(template string, vars map[string]string) (string, error) {
	// 占位符必须符合变量命名规则：字母开头，字母/数字/下划线
	re := regexp.MustCompile(`\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)
	matches := re.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		placeholder := match[0] // {变量名}
		varName := match[1]     // 变量名
		value, exists := vars[varName]
		if !exists {
			return "", errors.New("缺少变量: " + varName)
		}

		// 转义变量值，保证 JSON 合法
		escapedValue, err := json.Marshal(value)
		if err != nil {
			return "", fmt.Errorf("变量 %s 转义失败: %v", varName, err)
		}

		// 替换占位符（去掉 json.Marshal 返回的双引号）
		template = strings.ReplaceAll(template, placeholder, string(escapedValue[1:len(escapedValue)-1]))
	}

	return template, nil
}

// 通用函数组装（缺少变量跳过，不报错）
func SystemPromptComposeStatic(template string, vars map[string]string) (string, error) {
	re := regexp.MustCompile(`\{(\w+)\}`)
	matches := re.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		placeholder := match[0]
		varName := match[1]

		value, exists := vars[varName]
		if !exists {
			// 如果缺少变量，跳过，不报错
			continue
		}

		escapedValue, err := json.Marshal(value)
		if err != nil {
			return "", fmt.Errorf("变量 %s 转义失败: %v", varName, err)
		}

		template = strings.ReplaceAll(template, placeholder, string(escapedValue[1:len(escapedValue)-1]))
	}
	return template, nil
}

// openai reponse 2 json 解析函数
/*
func Response2JSON(resp string) (map[string]interface{}, error) {
	var m map[string]interface{}

	// 尝试严格解析
	if err := json.Unmarshal([]byte(resp), &m); err == nil {
		return m, nil
	}

	// 非严格解析，用正则匹配可能的 JSON，包括换行
	re := regexp.MustCompile(`(?s)\{.*\}`)
	match := re.FindString(resp)
	if match != "" {
		if err := json.Unmarshal([]byte(match), &m); err == nil {
			return m, nil
		}
	}

	// 都解析失败，返回一个空的 map，并返回一个解析错误
	return map[string]interface{}{}, errors.New("failed to parse response as JSON")
}
*/
// ============================  改进的 Response2JSON 函数
func Response2JSON(resp string) (map[string]interface{}, error) {
	var m map[string]interface{}

	// 1. 尝试严格解析
	if err := json.Unmarshal([]byte(resp), &m); err == nil {
		return m, nil
	}

	// 2. 清理响应文本，移除可能的非JSON前缀和后缀
	cleaned := cleanResponse(resp)
	if cleaned != "" {
		if err := json.Unmarshal([]byte(cleaned), &m); err == nil {
			return m, nil
		}
	}

	// 3. 尝试提取最可能的JSON对象
	jsonStr := extractMostLikelyJSON(resp)
	if jsonStr != "" {
		if err := json.Unmarshal([]byte(jsonStr), &m); err == nil {
			return m, nil
		}
	}

	// 4. 最后尝试：如果响应看起来像JSON但格式有问题，尝试修复
	fixedJSON := tryFixJSON(resp)
	if fixedJSON != "" {
		if err := json.Unmarshal([]byte(fixedJSON), &m); err == nil {
			return m, nil
		}
	}

	// 都解析失败，返回一个空 map，并返回一个解析错误
	return map[string]interface{}{}, errors.New("failed to parse response as JSON")
}

// 清理响应文本
func cleanResponse(resp string) string {
	// 移除常见的非JSON前缀和后缀
	prefixes := []string{
		"===== rawText =====",
		"生成json结果:",
		"当前时间:",
	}
	suffixes := []string{
		"===================",
	}

	cleaned := resp
	for _, prefix := range prefixes {
		if idx := strings.Index(cleaned, prefix); idx != -1 {
			cleaned = cleaned[idx+len(prefix):]
		}
	}

	for _, suffix := range suffixes {
		if idx := strings.Index(cleaned, suffix); idx != -1 {
			cleaned = cleaned[:idx]
		}
	}

	// 移除前后的空白字符
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}

// 提取最可能的JSON对象
func extractMostLikelyJSON(resp string) string {
	// 首先处理转义的换行符，将其转换为实际换行符
	resp = strings.ReplaceAll(resp, `\n`, "\n")

	// 匹配包含花括号的JSON对象，包括空对象
	// 使用更灵活的正则表达式来匹配各种格式的JSON
	re := regexp.MustCompile(`(?s)\{[\s\S]*?\}`)
	matches := re.FindAllString(resp, -1)

	if len(matches) == 0 {
		return ""
	}

	// 如果只有一个匹配项，直接返回
	if len(matches) == 1 {
		return matches[0]
	}

	// 优先选择包含最多键值对的匹配项
	var bestMatch string
	maxKeys := 0

	for _, match := range matches {
		// 计算大概的键值对数量
		keyCount := strings.Count(match, ":")
		if keyCount > maxKeys {
			maxKeys = keyCount
			bestMatch = match
		}
	}

	// 如果所有匹配项都是空对象，返回第一个
	if maxKeys == 0 && len(matches) > 0 {
		return matches[0]
	}

	return bestMatch
}

func tryFixJSON(resp string) string {
	// 移除可能的尾随逗号
	fixed := regexp.MustCompile(`,(\s*[}\]])`).ReplaceAllString(resp, "$1")

	// 确保字符串被正确引用
	fixed = regexp.MustCompile(`([{,]\s*)(\w+)(\s*:)`).ReplaceAllString(fixed, `${1}"${2}"${3}`)

	// 移除可能的注释
	fixed = regexp.MustCompile(`//.*`).ReplaceAllString(fixed, "")
	fixed = regexp.MustCompile(`/\*.*?\*/`).ReplaceAllString(fixed, "")

	return fixed
}

// ============================== ================================
// 随机生成 uuid
func GenerateUUID() string {
	return uuid.New().String()
}

// 解析Messages2text
func ParseMessages2Text(jsonStr string) (string, error) {
	var messages []Message
	if err := json.Unmarshal([]byte(jsonStr), &messages); err != nil {
		return "", err
	}

	var result string
	for _, msg := range messages {
		// 可根据 role 格式化
		switch msg.Role {
		case "user":
			result += "User: " + msg.Content + "\n"
		case "assistant", "system":
			// system 可以选择跳过或者也输出
			if msg.Role == "assistant" {
				result += "Assistant: " + msg.Content + "\n"
			}
			if msg.Role == "system" {
				result += "System: " + msg.Content + "\n"
			}
		default:
			// 其他角色直接输出,首字母大写
			result += strings.Title(msg.Role) + ": " + msg.Content + "\n"
		}
	}
	return result, nil
}

// ------------- 将Messages类解析成文本，用户画像专用 ---------------
// MessagesToText 将消息列表解析成文本
// 并按照User: Assistant: 进行分组
/*
User:
    user_content1  time1
    user_content2  time2
Assistant:
    assistant_content1  time1
    assistant_content2  time2



*/

func MessagesToText(messages []Message) string {
	var sb strings.Builder
	sb.WriteString("\n")
	for _, msg := range messages {
		sb.WriteString(msg.Role)
		sb.WriteString(": ")
		sb.WriteString(msg.Content)
		sb.WriteString("\n")
	}
	return sb.String()
}

// 结构体转文本
func Struct2JSON(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// 格式化时间戳
func FormatTimestamp(ts int64) string {
	return time.Unix(ts, 0).UTC().Format("2006-01-02 15:04:05")
}

// ConversationsToText 将对话对列表转换为文本，保留时间戳信息

// ------------- 类解析成文本，用户画像专用 ---------------
//  将消息列表解析成文本
// 并按照User: Assistant: 进行分组
/*
User:
    user_content1  time1
    user_content2  time2
Assistant:
    assistant_content1  time1
    assistant_content2  time2



*/
func ConversationsToText(conversations []Conversation) string {
	var sb strings.Builder

	// 分组存储
	userMsgs := []string{}
	assistantMsgs := []string{}

	for _, conv := range conversations {
		timestampStr := FormatTimestamp(conv.Timestamp)

		for _, msg := range conv.Messages {
			line := fmt.Sprintf("%s  %s", msg.Content, timestampStr)
			if msg.Role == "user" || msg.Role == "User" {
				userMsgs = append(userMsgs, line)
			} else if msg.Role == "assistant" || msg.Role == "Assistant" {
				assistantMsgs = append(assistantMsgs, line)
			}
		}
	}

	// 输出 User 部分
	sb.WriteString("User:\n")
	for _, m := range userMsgs {
		sb.WriteString("    ")
		sb.WriteString(m)
		sb.WriteString("\n")
	}

	// 输出 Assistant 部分
	sb.WriteString("Assistant:\n")
	for _, m := range assistantMsgs {
		sb.WriteString("    ")
		sb.WriteString(m)
		sb.WriteString("\n")
	}
	fmt.Printf("⌛️ process user and assistant messages events : %s", sb.String())

	return sb.String()
}

// ParseTimestamp 尝试解析多种时间格式
// ParseTimestamp 将时间字符串解析为 UTC 时间
func ParseTimestamp(tsStr string) (time.Time, error) {
	// 处理空字符串
	if tsStr == "" {
		return time.Time{}, fmt.Errorf("时间字符串为空")
	}

	// 去除前后空格
	tsStr = strings.TrimSpace(tsStr)

	// 尝试解析相对时间（如"昨天"、"明天"等），如果有实现 parseRelativeTime
	if relativeTime, err := parseRelativeTime(tsStr); err == nil {
		return relativeTime.UTC(), nil
	}

	formats := []string{
		"2006-01-02 15:04:05", // 完整时间
		"2006-01-02 15:04",    // 没有秒
		"2006-01-02",          // 只有日期
		"2006/01/02 15:04:05", // 带斜杠的完整时间
		"2006/01/02 15:04",    // 带斜杠无秒
		"2006/01/02",          // 带斜杠日期
		"2006.01.02 15:04:05", // 带点的完整时间
		"2006.01.02 15:04",    // 带点无秒
		"2006.01.02",          // 带点日期
		"15:04:05",            // 只有时间
		"15:04",               // 只有时间（无秒）
		time.RFC3339,          // ISO 8601
		time.RFC3339Nano,
		"2006-01-02T15:04:05", // ISO 8601 简化版
		"2006-01-02T15:04",    // ISO 8601 简化版（无秒）
	}

	var t time.Time
	var err error

	for _, f := range formats {
		// 优先用 Parse，能解析带时区的
		t, err = time.Parse(f, tsStr)
		if err == nil {
			return t.UTC(), nil
		}

		// 不带时区 → 按本地时区解析
		t, err = time.ParseInLocation(f, tsStr, time.Local)
		if err == nil {
			return t.UTC(), nil
		}
	}

	// 补全缺失秒数
	if len(tsStr) == len("2006-01-02 15:04") {
		tsStr += ":00"
		t, err = time.ParseInLocation("2006-01-02 15:04:05", tsStr, time.Local)
		if err == nil {
			return t.UTC(), nil
		}
	}

	// 只有时间 → 补全日期
	if len(tsStr) <= len("15:04:05") && strings.Contains(tsStr, ":") {
		currentDate := time.Now().Format("2006-01-02")
		fullTimeStr := currentDate + " " + tsStr
		t, err = time.ParseInLocation("2006-01-02 15:04:05", fullTimeStr, time.Local)
		if err == nil {
			return t.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("无法解析时间: %s", tsStr)
}

// parseRelativeTime 示例（根据需要实现）
func parseRelativeTime(s string) (time.Time, error) {
	// 简单示例，支持 "昨天"、"明天"
	now := time.Now()
	switch s {
	case "昨天":
		return now.AddDate(0, 0, -1), nil
	case "明天":
		return now.AddDate(0, 0, 1), nil
	default:
		return time.Time{}, fmt.Errorf("不是相对时间")
	}
}

func main() {
	tests := []string{
		"2025-09-26 16:42:12",
		"2025-09-26T16:42:12Z",
		"2025-09-26T16:42:12+08:00",
		"15:04:05",
		"昨天",
	}

	for _, ts := range tests {
		t, err := ParseTimestamp(ts)
		if err != nil {
			fmt.Println("错误:", err)
		} else {
			fmt.Println("原始:", ts, "→ UTC:", t.Format(time.RFC3339))
		}
	}
}
