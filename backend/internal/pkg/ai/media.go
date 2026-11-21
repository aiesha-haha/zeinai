package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

// MediaClient talks to the OpenAI Images and Videos (Sora) REST APIs directly.
//
// Environment:
//   OPENAI_API_KEY      - required for both images and videos
//   OPENAI_IMAGE_MODEL  - default "gpt-image-1" (also works with "dall-e-3")
//   OPENAI_VIDEO_MODEL  - default "sora-2"
type MediaClient struct {
	apiKey     string
	baseURL    string
	imageModel string
	videoModel string
	http       *http.Client
}

func NewMediaClient(apiKey string) *MediaClient {
	imageModel := os.Getenv("OPENAI_IMAGE_MODEL")
	if imageModel == "" {
		imageModel = "gpt-image-1"
	}
	videoModel := os.Getenv("OPENAI_VIDEO_MODEL")
	if videoModel == "" {
		videoModel = "sora-2"
	}
	return &MediaClient{
		apiKey:     apiKey,
		baseURL:    "https://api.openai.com/v1",
		imageModel: imageModel,
		videoModel: videoModel,
		http:       &http.Client{Timeout: 5 * time.Minute},
	}
}

// Configured reports whether an API key is present.
func (c *MediaClient) Configured() bool { return c != nil && c.apiKey != "" }

// apiError extracts a readable message from an OpenAI error response.
func apiError(status int, body []byte) error {
	var e struct {
		Error struct {
			Message string `json:"message"`
			Code    any    `json:"code"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &e) == nil && e.Error.Message != "" {
		return fmt.Errorf("openai %d: %s", status, e.Error.Message)
	}
	snippet := string(body)
	if len(snippet) > 300 {
		snippet = snippet[:300]
	}
	return fmt.Errorf("openai %d: %s", status, snippet)
}

func (c *MediaClient) do(req *http.Request) ([]byte, error) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, apiError(resp.StatusCode, body)
	}
	return body, nil
}

// GenerateImagePNG creates one image and returns its bytes.
func (c *MediaClient) GenerateImagePNG(ctx context.Context, prompt, size string) ([]byte, error) {
	if !c.Configured() {
		return nil, fmt.Errorf("OPENAI_API_KEY is not set")
	}
	if size == "" {
		size = "1024x1024"
	}
	payload := map[string]any{
		"model":  c.imageModel,
		"prompt": prompt,
		"size":   size,
		"n":      1,
	}
	// dall-e models need an explicit format; gpt-image models always return base64
	if strings.HasPrefix(c.imageModel, "dall-e") {
		payload["response_format"] = "b64_json"
	}
	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/images/generations", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	body, err := c.do(req)
	if err != nil {
		return nil, err
	}

	var res struct {
		Data []struct {
			B64JSON string `json:"b64_json"`
			URL     string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("invalid image response: %w", err)
	}
	if len(res.Data) == 0 {
		return nil, fmt.Errorf("image API returned no data")
	}
	if res.Data[0].B64JSON != "" {
		return base64.StdEncoding.DecodeString(res.Data[0].B64JSON)
	}
	if res.Data[0].URL != "" {
		dl, err := http.NewRequestWithContext(ctx, http.MethodGet, res.Data[0].URL, nil)
		if err != nil {
			return nil, err
		}
		resp, err := c.http.Do(dl)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			return nil, fmt.Errorf("image download failed: %d", resp.StatusCode)
		}
		return io.ReadAll(resp.Body)
	}
	return nil, fmt.Errorf("image API returned empty image")
}

// VideoJob is the state of a Sora video job.
type VideoJob struct {
	ID       string `json:"id"`
	Status   string `json:"status"` // queued | in_progress | completed | failed
	Progress int    `json:"progress"`
	Error    *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// StartVideo creates a video job. seconds: "4", "8" or "12"; size e.g. "1280x720".
func (c *MediaClient) StartVideo(ctx context.Context, prompt, seconds, size string) (*VideoJob, error) {
	if !c.Configured() {
		return nil, fmt.Errorf("OPENAI_API_KEY is not set")
	}
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("model", c.videoModel)
	_ = w.WriteField("prompt", prompt)
	if seconds != "" {
		_ = w.WriteField("seconds", seconds)
	}
	if size != "" {
		_ = w.WriteField("size", size)
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/videos", &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	body, err := c.do(req)
	if err != nil {
		return nil, err
	}
	var job VideoJob
	if err := json.Unmarshal(body, &job); err != nil {
		return nil, fmt.Errorf("invalid video response: %w", err)
	}
	if job.ID == "" {
		return nil, fmt.Errorf("video API returned no job id")
	}
	return &job, nil
}

// GetVideo returns the current state of a video job.
func (c *MediaClient) GetVideo(ctx context.Context, id string) (*VideoJob, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/videos/"+id, nil)
	if err != nil {
		return nil, err
	}
	body, err := c.do(req)
	if err != nil {
		return nil, err
	}
	var job VideoJob
	if err := json.Unmarshal(body, &job); err != nil {
		return nil, fmt.Errorf("invalid video status response: %w", err)
	}
	return &job, nil
}

// DownloadVideo streams the finished MP4 into dst.
func (c *MediaClient) DownloadVideo(ctx context.Context, id string, dst io.Writer) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/videos/"+id+"/content", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return apiError(resp.StatusCode, body)
	}
	_, err = io.Copy(dst, resp.Body)
	return err
}
