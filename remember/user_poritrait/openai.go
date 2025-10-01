package user_poritrait

import (
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

// 全局变量，直接暴露
var (
	LLMModel     string
	OpenAIClient openai.Client
)

func init() {
	InitLLM()
}

// InitLLM 初始化 OpenAI Client 并设置模型名称
func InitLLM() {
	OpenAIClient = openai.NewClient(
		option.WithAPIKey(Config.LLM.APIKey),
		option.WithBaseURL(Config.LLM.BaseURL),
	)
	LLMModel = Config.LLM.ModelID
}
