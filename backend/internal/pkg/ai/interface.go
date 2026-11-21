package ai

import (
	"context"
)

// TextGenerator is the abstraction for AI text generation
type TextGenerator interface {
	GenerateJSON(ctx context.Context, prompt string, target interface{}) error
	GenerateText(ctx context.Context, prompt string) (string, error)
}

// ImageGenerator is the abstraction for AI image generation
type ImageGenerator interface {
	GenerateImage(ctx context.Context, prompt string) (string, error)
}
