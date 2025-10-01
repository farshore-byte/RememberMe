package server

// 通用组装函数，支持变量插入不同的系统提示词模版，返回组装过后的系统提示词
import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
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

// ------------- 将Messages类解析成文本 ---------------
// MessagesToText 将消息列表解析成文本
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

// RAKE算法提取关键词
// ExtractKeywords RAKE 算法提取英文关键词
func ExtractKeywords(text string) []string {
	if text == "" {
		return []string{}
	}

	// 从文件加载停用词
	stopWords := loadStopWords("stopwords.txt")

	// 提取候选短语
	candidates := extractCandidateKeywords(text, stopWords)

	// 计算词频和共现频率
	wordFreq := make(map[string]int)
	wordDegree := make(map[string]int)
	wordCooccurrence := make(map[string]map[string]bool)

	for _, phrase := range candidates {
		words := strings.Fields(phrase)
		if len(words) > 5 { // 过长短语跳过共现计算，防止 O(n²)
			continue
		}

		uniqueWords := make(map[string]bool)
		for _, word := range words {
			word = strings.ToLower(word)
			wordFreq[word]++
			uniqueWords[word] = true
		}

		for w1 := range uniqueWords {
			for w2 := range uniqueWords {
				if w1 != w2 {
					if wordCooccurrence[w1] == nil {
						wordCooccurrence[w1] = make(map[string]bool)
					}
					wordCooccurrence[w1][w2] = true
				}
			}
		}
	}

	// 计算度
	for word := range wordFreq {
		degree := wordFreq[word]
		if co, ok := wordCooccurrence[word]; ok {
			degree += len(co)
		}
		wordDegree[word] = degree
	}

	// 计算短语分数
	keywordScores := make(map[string]float64)
	for _, phrase := range candidates {
		score := 0.0
		for _, word := range strings.Fields(phrase) {
			w := strings.ToLower(word)
			if wordDegree[w] > 0 {
				score += float64(wordDegree[w]) / float64(wordFreq[w])
			}
		}
		keywordScores[phrase] = score
	}

	// 排序
	type kv struct {
		Key   string
		Value float64
	}
	var sorted []kv
	for k, v := range keywordScores {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	// 返回前10个关键词
	topN := 10
	result := make([]string, 0, topN)
	for i, kv := range sorted {
		if i >= topN {
			break
		}
		result = append(result, kv.Key)
	}

	return result
}

// loadStopWords 从文件加载停用词
func loadStopWords(filename string) map[string]bool {
	stopWords := make(map[string]bool)

	// 获取当前源文件目录
	_, currentFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(currentFile)
	fullPath := filepath.Join(dir, filename)

	content, err := os.ReadFile(fullPath)
	if err != nil {
		// 文件不存在，返回默认停用词
		return getDefaultStopWords()
	}

	for _, line := range strings.Split(string(content), "\n") {
		word := strings.TrimSpace(strings.ToLower(line))
		if word != "" {
			stopWords[word] = true
		}
	}
	return stopWords
}

// extractCandidateKeywords 提取候选短语
func extractCandidateKeywords(text string, stopWords map[string]bool) []string {
	text = strings.ToLower(text)

	// 按句子分割
	sentences := strings.FieldsFunc(text, func(r rune) bool {
		return strings.ContainsRune(".!?;。！？；", r)
	})

	var candidates []string
	re := regexp.MustCompile(`[a-zA-Z0-9]+`) // 匹配英文单词或数字

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		words := re.FindAllString(sentence, -1)
		var phrase []string

		for _, w := range words {
			word := strings.ToLower(w)
			if stopWords[word] {
				if len(phrase) > 0 {
					candidates = append(candidates, strings.Join(phrase, " "))
					phrase = []string{}
				}
			} else {
				phrase = append(phrase, word)
			}
		}
		if len(phrase) > 0 {
			candidates = append(candidates, strings.Join(phrase, " "))
		}
	}

	return candidates
}

func getDefaultStopWords() map[string]bool {
	words := []string{
		"a", "an", "the", "and", "or", "but", "if", "while", "is", "are", "was", "were",
		"of", "at", "by", "for", "with", "about", "against", "between", "into", "through",
		"during", "before", "after", "above", "below", "to", "from", "up", "down", "in",
		"out", "on", "off", "over", "under", "again", "further", "then", "once", "here",
		"there", "when", "where", "why", "how", "all", "any", "both", "each", "few",
		"more", "most", "other", "some", "such", "no", "nor", "not", "only", "own",
		"same", "so", "than", "too", "very", "can", "will", "just", "don", "should", "now",
	}

	stopWords := make(map[string]bool, len(words))
	for _, w := range words {
		stopWords[w] = true
	}
	return stopWords
}

// 工具函数：返回 JSON
func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// GenerateSessionID 根据 group_id, user_id, role_id 生成 session_id
func GenerateSessionID(groupID, userID, roleID string) (string, error) {
	parts := []string{}
	if groupID != "" {
		parts = append(parts, groupID)
	}
	if userID != "" {
		parts = append(parts, userID)
	}
	if roleID != "" {
		parts = append(parts, roleID)
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("无法生成 session_id，group_id, user_id, role_id 全为空")
	}
	return strings.Join(parts, "_"), nil
}
