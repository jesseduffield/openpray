package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type openAIProvider struct {
	apiKey string
}

func (o *openAIProvider) Name() string { return "openai" }

type openAIRequest struct {
	Model               string          `json:"model"`
	Messages            []openAIMessage `json:"messages"`
	MaxCompletionTokens int64           `json:"max_completion_tokens"`
	// ReasoningEffort is set to "minimal" for reasoning models so they don't
	// spend the whole token budget thinking about the prayer instead of
	// praying.
	ReasoningEffort string `json:"reasoning_effort,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int64 `json:"prompt_tokens"`
		CompletionTokens int64 `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (o *openAIProvider) Complete(ctx context.Context, model, system, user string, maxTokens int64) (*Completion, error) {
	reqStruct := openAIRequest{
		Model: model,
		Messages: []openAIMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		MaxCompletionTokens: maxTokens,
	}
	if strings.HasPrefix(model, "gpt-5") {
		reqStruct.ReasoningEffort = "minimal"
	}
	body, err := json.Marshal(reqStruct)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("openai: reading response: %w", err)
	}

	var parsed openAIResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("openai: HTTP %d: %s", resp.StatusCode, truncate(string(data), 200))
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("openai: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("openai: empty response")
	}
	if parsed.Choices[0].Message.Content == "" {
		return nil, fmt.Errorf(
			"openai: model returned an empty prayer (finish_reason=%s) — "+
				"likely spent the whole budget reasoning; raise max_tokens",
			parsed.Choices[0].FinishReason)
	}

	return &Completion{
		Text:         parsed.Choices[0].Message.Content,
		InputTokens:  parsed.Usage.PromptTokens,
		OutputTokens: parsed.Usage.CompletionTokens,
	}, nil
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
