package chat_event

// --------------------  事件抽取 ------------------------------------

import (
	"fmt"
)

// ----------------------------------- 基础事件抽取模版 ----------------------------------

var ChatEventPromptTemplate = `
## 角色
你是梳理、抽取「关键事件」的专家。你的任务是抽根据「历史对话」，找出关键的时间节点及事件，并为其总结一句话，依次生成一条带有事件的时间线。

## 任务描述
「关键事件」时主人公user和assistant在角色扮演游戏中产生的对话提取到的，你需要准确识别「历史对话」中被规则定义为关键的事件，组织成具有先后顺序，符合格式规范的事件时间线。

## 任务约束
- 请不要大段摘抄对话中的内容，善于压缩成一句话
- 压缩事件经过时，明确你的事件描述能回答以下几个问题：
     - 谁
     - 谁干了什么
     务必确定好主体、行动、目标三要素
- 每一个事件视为一个条目，并以要求的格式输出
- 每个条目保持统一、清晰的写作风格，确保每个条目要素充分，先后顺序清晰
- 如果提取的事件中带有时间相关信息，你需要基于当前时间，推理出事件的时间，如上周三、昨天、明天等时间信息，并转化为具体日期
- 不在关键事件范围内的事件请不要记录，并跳过该时间点

## 关键事件范围
{event_scope}

只提取以上关键事件，其他事件不记录。

## 关键事件描述要求
1.语言具有描述行。输出结果会展示给终端用户，请保持语言表达的自然、流畅和简洁

## 当前时间
{current_time}

## 历史对话
{messages_str}

#输出格式示例
要求输出json格式，key值为时间，需要满足时间格式解析规则，value为关键事件。
注意：如果没有任何关键事件，则输出为空json。


{format_example}



# 输出示例1

{output_example_1}


# 语言设置
- 你的工作只有{language}

# 生成json结果
`

// ---------------------------------------- 业务变量 ---------------------------------------------

// 事件范围
var EventScope = `1.结伴而行的活动
2.用户向角色提供物品
3.用户向角色提供奖励
4.用户向角色提供惩罚
5.角色向用户提供物品
6.角色向用户提供奖励
7.角色向用户提供惩罚
8.约定条目的键 必须以约定生效的时间节点填写。例如: 2025-09-13 12：30 :user made an appointment to go for a walk together on 2025-09-13 12:30
9.当有第三者介入时发生的事件`

// 输出格式示例
var ChatEventFormatExample = `
当前时间: XXXX-XX-XX XX:XX 星期X
对话中发生的关键事件有 ["user_name和role_name一起做爱","user_name和role_name约定好2025-09-13见面"]
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
对话中发生的关键事件有 ["Lisa给kim做早餐","kim和lisa一起玩真心话大冒险","lisa输了，给kim跳了一支舞蹈","kim约定和Lisa在 2025-09-13 在茶餐厅吃个饭"]
生成json结果：
{{
 "2025-09-11 08:30" : "Lisa给user做早餐",
 "2025-09-11 12:30" : "user和lisa一起玩真心话大冒险",
 "2025-09-11 14:00" : "lisa输了，给kim跳了一支舞蹈",
 "2025-09-13 12:30" : "kim约定和Lisa在 2025-09-13 茶餐厅吃个饭"
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
