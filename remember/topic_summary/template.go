package topic_summary

// 提示词建议使用redis存储，前期写在代码里调试
import (
	"fmt"
)

var TopicSummaryPromptTemplate = `
# Conversation Topic Summary System Prompt

## Role
You are an expert in topic extraction. Your task is to generate a sentence that corresponds to the topic based on the provided "historical conversation."

## Task Description
Extract topic-related content from a user or character's short speech. Long descriptions, actions, and events are prohibited.

## Restrictions
- Topics must be **single words or phrases**; do not copy long paragraphs.
- First, extract the topics mentioned in the conversation.
- For each topic, extract one or more phrases to support them.
- Discard irrelevant asides and descriptions from the character's responses.
- Each summary must:
- Be as concise as possible.
- Be relevant to the topic to which it belongs.
- Contain the strongest evidence from the original text that supports the topic.
- Entities (person names, book titles, places, brands, product categories, concepts). **The original text must be referenced in parentheses**.

## Topic Guidelines
{topics_str}

## Historical Conversations
{messages_str}

## Output Format
- Output must be in **JSON** format. - Key = Topic (strictly adheres to topic constraints)
- Value = One-sentence summary.

### Example:

{output_example_1}

## Language Settings
- All output uses only {language}.

# Topic Summary Results
`
var Topics = `1. [fitness, soccer, basketball, travel, music, writing, painting, gaming, …]  
2. [reading, programming, cooking, sex education, psychology, history, literature, …]  
3. [food, shopping, health, family, work, friends, …]  
4. [romance, flirtation, friendship, conflict, cooperation, …]  
5. [life goals, morality, future plans, beliefs, boundaries, …]`

var TopicOutputExample1 = `The current dialogue covers topics [fitness, soccer, reading, cooking].
I will generate one first-person summary sentence for each topic and organize them in JSON:

{
  "fitness": "Kim requested to maintain daily morning runs, so I set a weekly (four-on, three-off) running schedule and encouraged them to persist.",
  "soccer": "Kim asked whether there would be a soccer match tonight; I reminded them that the match would be at (8 PM tonight) between teams (Argentina) and (Netherlands).",
  "reading": "Kim wanted new novel recommendations, so I organized (The Old Man and the Sea) and (How the Steel Was Tempered) as suggested books.",
  "cooking": "Kim asked how to make Italian pasta; I provided detailed guidance on preparing (noodles), (Italian meat sauce), (lettuce), and the steps of (frying meat patties, boiling pasta, pouring sauce)."
}`

var TopicLanguage = "English"

//-------------------------------- 代码 -------------------------------------

var TopicStaticVars = map[string]string{
	"topics_str":       Topics,
	"output_example_1": TopicOutputExample1,
	"language":         TopicLanguage,
}

type TopicTemplat struct {
	Template   string            // 静态变量已替换后的模板
	StaticVars map[string]string // 静态变量
}

func NewTopicSummaryTemplate() *TopicTemplat {
	return &TopicTemplat{
		Template:   TopicSummaryPromptTemplate,
		StaticVars: TopicStaticVars,
	}
}
func (t *TopicTemplat) BuildPrompt(dynamicVars map[string]string) (string, error) {
	// 先替换静态变量
	finalTpl, err := SystemPromptComposeStatic(t.Template, t.StaticVars)
	if err != nil {
		return "", fmt.Errorf("静态模板组装失败: %v", err)
	}

	// 再替换动态变量
	finalTpl, err = SystemPromptCompose(finalTpl, dynamicVars)
	if err != nil {
		return "", fmt.Errorf("动态模板组装失败: %v", err)
	}

	return finalTpl, nil
}
