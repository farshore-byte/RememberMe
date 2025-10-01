package chat_event

// --------------------  事件抽取 ------------------------------------

import (
	"fmt"
)

// ----------------------------------- 基础事件抽取模版 ----------------------------------

var ChatEventPromptTemplate = `
## Role

You are an expert at sorting and extracting "minimal event units." Your task is to extract the most concise event units from historical conversations.

## Minimal Event Unit Definition:
- It should condense the information contained in the text to the maximum possible extent.
- Events should be as concise as possible.
- Events should not contain descriptions or narration.
- Events should not contain dialogue.
- Each event should be treated as a single entry and output in a concise format.
- Extract events from user or role groups.
- If the extracted events contain time information, you should infer the event time based on the current time (e.g., last Wednesday, yesterday, or tomorrow) and convert it to a specific date.
- Events that do not fall within the critical event scope should not be recorded and should be skipped.

## Minimal Event Scope
{event_scope}

Extract only the minimal events listed above; other events will not be recorded.

## Minimum Event Requirements

Person  - Action - Place  - Time - Result

For example:

Lisa made breakfast for Kim

The user scheduled a walk with Lisa on 2025-09-11 at 12:30 PM

## Current time
{current_time}

## Historical conversations
{messages_str}

## Output format example
The output is in JSON format. The key should be a time that conforms to the time format parsing rules, and the value should be a key event.

Note: If there is no key event, the output will be an empty JSON file.

Event value output format example:
Person Action Place Time Result

Use spaces to separate columns.

{format_example}

# Output example 1

{output_example_1}

# Language settings
- Your task is to use {language}

# Generate JSON output
`

// ---------------------------------------- 业务变量 ---------------------------------------------

// 事件范围
var EventScope = `1. Walking together
2. User actions
2. The key for the appointment entry must be filled in with the time when the appointment takes effect. For example: 2025-09-13 12:30: Kim made an appointment to go for a walk together on 2025-09-13 12:30
6. User interactions with the real world (places, objects, props, etc.)`

// 输出格式示例
var ChatEventFormatExample = `
当前时间: XXXX-XX-XX XX:XX 星期X
生成json结果:
{{
 "2025-09-11" : ".......",
 "2025-09-11 12:30" : "......",
 "2025-09-11 14:00" : ".......",
 "2025-09-13 17:00" : "......." 
}}`

// 输出示例1
var ChatEventOutputExample1 = `
当前时间: 2025-09-11 12:30 星期一
生成json结果：
{{
 "2025-09-11 08:30" : "Lisa 给 user 做早餐",
 "2025-09-11 12:30" : "kim 和 lisa 一起玩 真心话大冒险",
 "2025-09-11 14:00" : "lisa 输了， 给 kim 跳了 舞蹈",
 "2025-09-13 12:30" : "kim 约定 Lisa 在 2025-09-13 茶餐厅 吃个饭"
}}
 `

// 语言设置
var ChatEventLanguage = `英文`

//------------------------------------ 代码 -------------------------------------------

var Template *ChatEventTemplate

func init() {
	Template = NewChatEventTemplate()
}

var ChatEventtaticVars = map[string]string{
	"event_scope":      EventScope,
	"format_example":   ChatEventFormatExample,
	"output_example_1": ChatEventOutputExample1,
	"language":         ChatEventLanguage,
}

// 动态变量类
type ChatEventDynamicVars struct {
	CurrentTime string // 对应 "{current_time}"
	MessagesStr string // 对应 "{messages_str}"，将 []string join 成字符串
}

type ChatEventTemplate struct {
	Template   string            // 静态变量已替换后的模板
	StaticVars map[string]string // 静态变量
}

func NewChatEventTemplate() *ChatEventTemplate {
	return &ChatEventTemplate{
		Template:   ChatEventPromptTemplate,
		StaticVars: ChatEventtaticVars,
	}
}
func (c *ChatEventTemplate) BuildPrompt(dynamicVars *ChatEventDynamicVars) (string, error) {
	// 先替换静态变量
	finalTpl, err := SystemPromptComposeStatic(c.Template, c.StaticVars)
	if err != nil {
		return "", fmt.Errorf("静态模板组装失败: %v", err)
	}

	// 将动态变量转成 map[string]string
	dynMap := map[string]string{
		"current_time": dynamicVars.CurrentTime,
		"messages_str": dynamicVars.MessagesStr,
	}

	// 替换动态变量
	finalTpl, err = SystemPromptCompose(finalTpl, dynMap)
	if err != nil {
		return "", fmt.Errorf("动态模板组装失败: %v", err)
	}

	return finalTpl, nil
}
