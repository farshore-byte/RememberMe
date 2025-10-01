package topic_summary

import (
	"context"
	"fmt"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

// 输入参数
type ExecuteRequest struct {
	SystemPrompt string
	Query        string
	Client       *openai.Client
	Model        string
}

// 执行结果
type ExecuteResult struct {
	JSON map[string]interface{} // 转换后的JSON结果
}

// 执行函数，只会输出json结果
func Execute(req *ExecuteRequest) (*ExecuteResult, error) {
	if req.Client == nil {
		return nil, fmt.Errorf("client is nil")
	}

	ctx := context.Background()
	resp, err := req.Client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(req.SystemPrompt),
				openai.UserMessage(req.Query),
			},
			Model:           req.Model,
			ReasoningEffort: "minimal",
		},
		option.WithJSONSet("thinking", map[string]string{
			"type": "disabled", // 禁用深度思考
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("openai request failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("empty response from model")
	}
	rawText := resp.Choices[0].Message.Content

	jsonResult, err := Response2JSON(rawText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response as JSON: %w", err)
	}

	return &ExecuteResult{
		JSON: jsonResult,
	}, nil
}
