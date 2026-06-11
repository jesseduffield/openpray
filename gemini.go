package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type geminiProvider struct {
	apiKey string
}

func (g *geminiProvider) Name() string { return "gemini" }

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	SystemInstruction *geminiContent  `json:"systemInstruction,omitempty"`
	Contents          []geminiContent `json:"contents"`
	GenerationConfig  struct {
		MaxOutputTokens int64 `json:"maxOutputTokens"`
	} `json:"generationConfig"`
}

type geminiResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int64 `json:"promptTokenCount"`
		CandidatesTokenCount int64 `json:"candidatesTokenCount"`
	} `json:"usageMetadata"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (g *geminiProvider) Complete(ctx context.Context, model, system, user string, maxTokens int64) (*Completion, error) {
	reqBody := geminiRequest{
		SystemInstruction: &geminiContent{Parts: []geminiPart{{Text: system}}},
		Contents: []geminiContent{
			{Role: "user", Parts: []geminiPart{{Text: user}}},
		},
	}
	reqBody.GenerationConfig.MaxOutputTokens = maxTokens

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent",
		url.PathEscape(model))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", g.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gemini: reading response: %w", err)
	}

	var parsed geminiResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("gemini: HTTP %d: %s", resp.StatusCode, truncate(string(data), 200))
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("gemini: %s", parsed.Error.Message)
	}
	if len(parsed.Candidates) == 0 {
		return nil, fmt.Errorf("gemini: empty response")
	}

	var text string
	for _, part := range parsed.Candidates[0].Content.Parts {
		text += part.Text
	}
	return &Completion{
		Text:         text,
		InputTokens:  parsed.UsageMetadata.PromptTokenCount,
		OutputTokens: parsed.UsageMetadata.CandidatesTokenCount,
	}, nil
}
