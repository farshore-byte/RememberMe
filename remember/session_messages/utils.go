package session_messages

// 通用组装函数，支持变量插入不同的系统提示词模版，返回组装过后的系统提示词
import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
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
func Response2JSON(resp string) (map[string]interface{}, error) {
	var m map[string]interface{}

	// 尝试严格解析
	if err := json.Unmarshal([]byte(resp), &m); err == nil {
		return m, nil
	}

	// 非严格解析，用正则匹配可能的 JSON，包括换行
	re := regexp.MustCompile(`(?s)\{.*?\}`)
	matches := re.FindAllString(resp, -1) // 获取所有匹配的 JSON

	if len(matches) > 0 {
		lastMatch := matches[len(matches)-1] // 取最后一个匹配
		if err := json.Unmarshal([]byte(lastMatch), &m); err == nil {
			return m, nil
		}
	}

	// 都解析失败，返回一个空 map，并返回一个解析错误
	return map[string]interface{}{}, errors.New("failed to parse response as JSON")
}

// formatMessagesToRoleContent 将 MemoryMessage 转换为 role-content 格式
func formatMessagesToRoleContent(messages []MemoryMessage) []map[string]string {
	result := []map[string]string{} // 初始化为空数组

	for _, msg := range messages {
		if msg.UserContent != "" {
			result = append(result, map[string]string{
				"role":       "user",
				"content":    msg.UserContent,
				"timestamp":  FormatTimestamp(msg.CreatedAt.Unix()),
				"created_at": msg.CreatedAt.UTC().Format(time.RFC3339), // 转成 UTC 字符串
			})
		}

		if msg.AssistantContent != "" {
			result = append(result, map[string]string{
				"role":       "assistant",
				"content":    msg.AssistantContent,
				"timestamp":  FormatTimestamp(msg.CreatedAt.Unix()),
				"created_at": msg.CreatedAt.UTC().Format(time.RFC3339), // 转成 UTC 字符串
			})
		}
	}

	return result
}

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

// processMessages 处理 messages 数组，过滤非支持字段并确保成对
func processMessages(messages []map[string]interface{}) ([]struct {
	UserContent      string
	AssistantContent string
}, error) {
	var result []struct {
		UserContent      string
		AssistantContent string
	}

	var currentUserContent string

	for _, msg := range messages {
		// 过滤非支持字段，只保留 role 和 content
		role, roleOk := msg["role"].(string)
		content, contentOk := msg["content"].(string)

		if !roleOk || !contentOk {
			continue // 跳过无效的消息
		}

		switch role {
		case "user":
			// 如果已经有未配对的用户消息，先保存它（助手内容为空）
			if currentUserContent != "" {
				result = append(result, struct {
					UserContent      string
					AssistantContent string
				}{
					UserContent:      currentUserContent,
					AssistantContent: "",
				})
			}
			currentUserContent = content

		case "assistant":
			// 将用户消息和助手消息配对
			result = append(result, struct {
				UserContent      string
				AssistantContent string
			}{
				UserContent:      currentUserContent,
				AssistantContent: content,
			})
			currentUserContent = "" // 重置当前用户消息

		default:
			// 忽略其他角色
			continue
		}
	}

	// 处理最后一个未配对的用户消息
	if currentUserContent != "" {
		result = append(result, struct {
			UserContent      string
			AssistantContent string
		}{
			UserContent:      currentUserContent,
			AssistantContent: "",
		})
	}

	return result, nil
}

// ------------- 将Messages类解析成文本 ---------------
// MessagesToText 将消息列表解析成文本
func MessagesToText(messages []Message) string {
	var sb strings.Builder
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
	return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
}
