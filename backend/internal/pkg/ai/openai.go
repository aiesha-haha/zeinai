package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"zeinai-backend/internal/config"

	"github.com/sashabaranov/go-openai"
)

type OpenAIGenerator struct {
	client *openai.Client
	model  string
}

func NewOpenAIGenerator(cfg *config.Config) *OpenAIGenerator {
	return &OpenAIGenerator{
		client: openai.NewClient(cfg.OpenAI.APIKey),
		model:  openai.GPT4oMini, // Default to a cost-effective but powerful model
	}
}

func (g *OpenAIGenerator) GenerateText(ctx context.Context, prompt string) (string, error) {
	resp, err := g.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: g.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("openai generation failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty openai response")
	}

	return resp.Choices[0].Message.Content, nil
}

func (g *OpenAIGenerator) GenerateJSON(ctx context.Context, prompt string, target interface{}) error {
	resp, err := g.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: g.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
		},
	)

	if err != nil {
		return fmt.Errorf("openai json generation failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return fmt.Errorf("empty openai response")
	}

	return json.Unmarshal([]byte(resp.Choices[0].Message.Content), target)
}

func (g *OpenAIGenerator) GenerateImage(ctx context.Context, prompt string) (string, error) {
	resp, err := g.client.CreateImage(
		ctx,
		openai.ImageRequest{
			Prompt:         prompt,
			Model:          openai.CreateImageModelDallE3,
			N:              1,
			Size:           openai.CreateImageSize1024x1024,
			ResponseFormat: openai.CreateImageResponseFormatURL,
		},
	)
	if err != nil {
		return "", fmt.Errorf("dall-e generation failed: %w", err)
	}

	if len(resp.Data) == 0 {
		return "", fmt.Errorf("empty dall-e response")
	}

	return resp.Data[0].URL, nil
}
