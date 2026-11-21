package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"zeinai-backend/internal/config"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiGenerator struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

func NewGeminiGenerator(ctx context.Context, cfg *config.Config) (*GeminiGenerator, error) {
	if cfg.Gemini.APIKey == "" {
		return nil, fmt.Errorf("gemini API key is missing")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.Gemini.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	model := client.GenerativeModel("gemini-1.5-flash")

	return &GeminiGenerator{
		client: client,
		model:  model,
	}, nil
}

func (g *GeminiGenerator) GenerateText(ctx context.Context, prompt string) (string, error) {
	resp, err := g.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("gemini generation failed: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty gemini response")
	}

	var parts []string
	for _, part := range resp.Candidates[0].Content.Parts {
		parts = append(parts, fmt.Sprintf("%v", part))
	}

	return strings.Join(parts, ""), nil
}

func (g *GeminiGenerator) GenerateJSON(ctx context.Context, prompt string, target interface{}) error {
	// Gemini 1.5 doesn't always strictly support JSON mode in the same way OpenAI does via SDK easily
	// (it has it, but requires more config). For now, use text + cleaning.
	text, err := g.GenerateText(ctx, prompt)
	if err != nil {
		return err
	}

	cleaned := cleanJSON(text)
	return json.Unmarshal([]byte(cleaned), target)
}

func cleanJSON(s string) string {
	s = strings.TrimSpace(s)
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start != -1 && end != -1 && end > start {
		return s[start : end+1]
	}
	return s
}

func (g *GeminiGenerator) Close() {
	if g.client != nil {
		g.client.Close()
	}
}
