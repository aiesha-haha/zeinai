package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// GeminiREST calls the Gemini generateContent REST API directly.
// It is used as a text fallback when Kimi / OpenAI are unavailable.
//
// Environment: GEMINI_API_KEY (required), GEMINI_MODEL (default "gemini-3.8-flash").
type GeminiREST struct {
	apiKey string
	model  string
	http   *http.Client
}

func NewGeminiREST(apiKey string) *GeminiREST {
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-3.8-flash"
	}
	return &GeminiREST{apiKey: apiKey, model: model, http: &http.Client{Timeout: 3 * time.Minute}}
}

func (g *GeminiREST) generate(ctx context.Context, prompt string, asJSON bool) (string, error) {
	body := map[string]any{
		"contents": []map[string]any{{"parts": []map[string]string{{"text": prompt}}}},
	}
	if asJSON {
		body["generationConfig"] = map[string]any{"responseMimeType": "application/json"}
	}
	b, _ := json.Marshal(body)
	url := "https://generativelanguage.googleapis.com/v1beta/models/" + g.model + ":generateContent"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", g.apiKey)

	resp, err := g.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var out struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.Unmarshal(raw, &out)
	if resp.StatusCode >= 300 {
		msg := string(raw)
		if out.Error != nil && out.Error.Message != "" {
			msg = out.Error.Message
		}
		// "status code: NNN" lets the resilient wrapper recognise 429 / 5xx as temporary
		return "", fmt.Errorf("gemini generation failed: status code: %d, %s", resp.StatusCode, msg)
	}
	text := ""
	if len(out.Candidates) > 0 {
		for _, p := range out.Candidates[0].Content.Parts {
			text += p.Text
		}
	}
	if text == "" {
		return "", fmt.Errorf("gemini returned an empty answer")
	}
	return text, nil
}

func (g *GeminiREST) GenerateText(ctx context.Context, prompt string) (string, error) {
	return g.generate(ctx, prompt, false)
}

func (g *GeminiREST) GenerateJSON(ctx context.Context, prompt string, target interface{}) error {
	text, err := g.generate(ctx, prompt, true)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(cleanJSON(text)), target); err != nil {
		return fmt.Errorf("gemini returned invalid JSON: %w", err)
	}
	return nil
}
