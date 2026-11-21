package ai

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"zeinai-backend/internal/config"
)

// ResilientGenerator wraps text providers so that a short rate-limit (HTTP 429
// "rate limit") is retried with a back-off, and a provider that keeps failing
// (no credits, invalid key, outage) is replaced by the next configured one.
type ResilientGenerator struct {
	names     []string
	providers []TextGenerator
	backoff   []time.Duration
}

type namedGenerator struct {
	name string
	gen  TextGenerator
	key  string
}

// NewResilientGenerator builds a chain of providers in the given priority.
// Providers without an API key are skipped.
func newResilient(chain ...namedGenerator) *ResilientGenerator {
	r := &ResilientGenerator{backoff: []time.Duration{2 * time.Second, 5 * time.Second, 12 * time.Second}}
	for _, c := range chain {
		if c.key == "" || c.gen == nil {
			continue
		}
		r.names = append(r.names, c.name)
		r.providers = append(r.providers, c.gen)
	}
	return r
}

// KimiFirst prefers Kimi, then OpenAI, then Gemini.
func KimiFirst(cfg *config.Config, kimi, openai TextGenerator) *ResilientGenerator {
	return newResilient(
		namedGenerator{"kimi", kimi, cfg.Kimi.APIKey},
		namedGenerator{"openai", openai, cfg.OpenAI.APIKey},
		namedGenerator{"gemini", NewGeminiREST(cfg.Gemini.APIKey), cfg.Gemini.APIKey},
	)
}

// OpenAIFirst prefers OpenAI, then Kimi, then Gemini.
func OpenAIFirst(cfg *config.Config, openai, kimi TextGenerator) *ResilientGenerator {
	return newResilient(
		namedGenerator{"openai", openai, cfg.OpenAI.APIKey},
		namedGenerator{"kimi", kimi, cfg.Kimi.APIKey},
		namedGenerator{"gemini", NewGeminiREST(cfg.Gemini.APIKey), cfg.Gemini.APIKey},
	)
}

// isRetryable reports a temporary rate limit (worth waiting for), as opposed to
// "no credits / quota exceeded" which will not fix itself in seconds.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	if strings.Contains(s, "credit") || strings.Contains(s, "quota") || strings.Contains(s, "billing") {
		return false
	}
	return strings.Contains(s, "429") || strings.Contains(s, "rate limit") ||
		strings.Contains(s, "status code: 5") || strings.Contains(s, "overloaded") ||
		strings.Contains(s, "high demand") || strings.Contains(s, "unavailable")
}

func (r *ResilientGenerator) run(ctx context.Context, call func(TextGenerator) error) error {
	if len(r.providers) == 0 {
		return fmt.Errorf("no AI text provider configured: set KIMI_API_KEY, OPENAI_API_KEY or GEMINI_API_KEY")
	}
	var errs []string
	for i, p := range r.providers {
		err := call(p)
		for attempt := 0; err != nil && isRetryable(err) && attempt < len(r.backoff); attempt++ {
			log.Printf("[AI] %s busy (%v), retry in %s", r.names[i], err, r.backoff[attempt])
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(r.backoff[attempt]):
			}
			err = call(p)
		}
		if err == nil {
			return nil
		}
		errs = append(errs, fmt.Sprintf("%s: %v", r.names[i], err))
		if i < len(r.providers)-1 {
			log.Printf("[AI] %s failed, switching to %s: %v", r.names[i], r.names[i+1], err)
		}
	}
	return fmt.Errorf("all AI providers failed: %s", strings.Join(errs, " | "))
}

func (r *ResilientGenerator) GenerateJSON(ctx context.Context, prompt string, target interface{}) error {
	return r.run(ctx, func(g TextGenerator) error { return g.GenerateJSON(ctx, prompt, target) })
}

func (r *ResilientGenerator) GenerateText(ctx context.Context, prompt string) (string, error) {
	var out string
	err := r.run(ctx, func(g TextGenerator) error {
		s, e := g.GenerateText(ctx, prompt)
		out = s
		return e
	})
	return out, err
}
