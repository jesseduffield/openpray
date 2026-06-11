package main

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
)

type anthropicProvider struct {
	client anthropic.Client
}

func newAnthropicProvider() *anthropicProvider {
	// Reads ANTHROPIC_API_KEY from the environment.
	return &anthropicProvider{client: anthropic.NewClient()}
}

func (a *anthropicProvider) Name() string { return "anthropic" }

func (a *anthropicProvider) Complete(ctx context.Context, model, system, user string, maxTokens int64) (*Completion, error) {
	resp, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: maxTokens,
		System: []anthropic.TextBlockParam{
			{Text: system},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(user)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("anthropic: %w", err)
	}

	var text string
	for _, block := range resp.Content {
		switch variant := block.AsAny().(type) {
		case anthropic.TextBlock:
			text += variant.Text
		}
	}
	return &Completion{
		Text:         text,
		InputTokens:  resp.Usage.InputTokens,
		OutputTokens: resp.Usage.OutputTokens,
	}, nil
}
