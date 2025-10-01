package user_poritrait

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

// 随机生成 uuid
func GenerateUUID() string {
	return uuid.New().String()
}

// 解析Messages2text， 专门服务于用户画像的提取
/*

过滤掉assistant、system 的content

*/
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
			// 过滤掉 assistant、system 的content
			continue
		default:
			// 其他角色同样过滤
			continue
		}
	}
	return result, nil
}

// MessagesToText 将消息列表解析成文本
// 类似的，
// MessagesToText 将消息列表解析成文本，只保留 user 消息
func MessagesToText(messages []Message) string {
	var sb strings.Builder
	sb.WriteString("\n")
	for _, msg := range messages {
		if msg.Role == "user" {
			sb.WriteString("User: ")
			sb.WriteString(msg.Content)
			sb.WriteString("\n")
		}
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
