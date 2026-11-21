package ai

import (
	"context"
	"fmt"
)

// MockGenerator implements TextGenerator with static/procedural responses for development
type MockGenerator struct{}

func NewMockGenerator() *MockGenerator {
	return &MockGenerator{}
}

func (g *MockGenerator) GenerateText(ctx context.Context, prompt string) (string, error) {
	return "This is a mock AI response because no valid API key was found. Please configure KIMI_API_KEY, OPENAI_API_KEY, or GEMINI_API_KEY.", nil
}

func (g *MockGenerator) GenerateJSON(ctx context.Context, prompt string, target interface{}) error {
	// We try to satisfy basic JSON structures for our workers
	// This is a naive implementation that should be used with care
	
	// For WhiteboardWorker
	if fmt.Sprintf("%T", target) == "*worker.whiteboardAIResponse" || fmt.Sprintf("%T", target) == "*main.whiteboardAIResponse" {
		t := target.(interface{}).(*struct {
			Summary     string `json:"summary"`
			Explanation string `json:"explanation"`
		})
		t.Summary = "### Mock Lesson Summary\n\n- The lesson data was captured successfully.\n- AI summary generation is currently disabled or missing a valid API key.\n- You can still view the original board."
		t.Explanation = "This is a placeholder explanation. To enable real AI summaries, please provide a valid Kimi or OpenAI API key in the .env file."
		return nil
	}

	return fmt.Errorf("mock generator: unsupported JSON target type %T", target)
}
