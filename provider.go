package main

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// Completion is one offering returned by a provider.
type Completion struct {
	Text         string
	InputTokens  int64
	OutputTokens int64
}

// Provider abstracts an LLM vendor. Bring your own API key via the
// environment; openpray never stores keys.
type Provider interface {
	Name() string
	Complete(ctx context.Context, model, system, user string, maxTokens int64) (*Completion, error)
}

func newProvider(name string) (Provider, error) {
	switch strings.ToLower(name) {
	case "anthropic":
		if os.Getenv("ANTHROPIC_API_KEY") == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY is not set")
		}
		return newAnthropicProvider(), nil
	case "openai":
		key := os.Getenv("OPENAI_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY is not set")
		}
		return &openAIProvider{apiKey: key}, nil
	case "gemini", "google":
		key := os.Getenv("GEMINI_API_KEY")
		if key == "" {
			key = os.Getenv("GOOGLE_API_KEY")
		}
		if key == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY is not set")
		}
		return &geminiProvider{apiKey: key}, nil
	default:
		return nil, fmt.Errorf("unknown provider %q (want anthropic, openai, or gemini)", name)
	}
}
