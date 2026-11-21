package ai

import (
	"os"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"zeinai-backend/internal/config"

	"github.com/sashabaranov/go-openai"
)

// KimiGenerator implements TextGenerator using Kimi K2 (OpenAI-compatible API)
type KimiGenerator struct {
	client *openai.Client
	model  string
}

// NewKimiGenerator creates a new Kimi K2 generator.
// Kimi uses an OpenAI-compatible API – we just point the base URL at Moonshot AI.
func NewKimiGenerator(cfg *config.Config) *KimiGenerator {
	clientCfg := openai.DefaultConfig(cfg.Kimi.APIKey)
	clientCfg.BaseURL = "https://api.moonshot.ai/v1"

	return &KimiGenerator{
		client: openai.NewClientWithConfig(clientCfg),
		model:  kimiModel(),
	}
}

// GenerateText sends a plain-text prompt and returns the response string.
func (g *KimiGenerator) GenerateText(ctx context.Context, prompt string) (string, error) {
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
		return "", fmt.Errorf("kimi generation failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty kimi response")
	}

	return resp.Choices[0].Message.Content, nil
}

// GenerateJSON sends a prompt expecting JSON back and unmarshals it into target.
func (g *KimiGenerator) GenerateJSON(ctx context.Context, prompt string, target interface{}) error {
	resp, err := g.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: g.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a helpful assistant that always responds with valid JSON only. Do not include markdown code blocks, explanations, or any text outside the JSON object.",
				},
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
		return fmt.Errorf("kimi json generation failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return fmt.Errorf("empty kimi response")
	}

	raw := resp.Choices[0].Message.Content
	cleaned := cleanKimiJSON(raw)
	if err := json.Unmarshal([]byte(cleaned), target); err != nil {
		return fmt.Errorf("kimi response json parse failed: %w (raw: %.200s)", err, raw)
	}
	return nil
}

// cleanKimiJSON strips optional markdown fences that Kimi may add.
func cleanKimiJSON(s string) string {
	s = strings.TrimSpace(s)
	// Strip ```json ... ``` fences if present
	if strings.HasPrefix(s, "```") {
		start := strings.Index(s, "\n")
		end := strings.LastIndex(s, "```")
		if start != -1 && end > start {
			s = strings.TrimSpace(s[start+1 : end])
		}
	}
	// Fallback: find first { and last }
	first := strings.Index(s, "{")
	last := strings.LastIndex(s, "}")
	if first != -1 && last != -1 && last > first {
		return s[first : last+1]
	}
	return s
}

// kimiModel lets the server choose the Moonshot model (KIMI_MODEL); older
// preview models get retired, so this avoids a code change when that happens.
func kimiModel() string {
	if m := os.Getenv("KIMI_MODEL"); m != "" {
		return m
	}
	return "kimi-k2-0711-preview"
}
