package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type GeminiProvider struct {
	Key     string
	ModelID string
	BaseURL string
}

func (p GeminiProvider) Name() string     { return "gemini" }
func (p GeminiProvider) Model() string    { return p.ModelID }
func (p GeminiProvider) Endpoint() string { return p.BaseURL }
func (p GeminiProvider) Local() bool      { return false }

func (p GeminiProvider) Complete(ctx context.Context, system, user string) (string, error) {
	endpoint := strings.TrimSuffix(p.BaseURL, "/") +
		"/v1beta/models/" + p.ModelID + ":generateContent?key=" + p.Key
	raw, err := postJSON(ctx, endpoint, "", "", nil, map[string]any{
		"system_instruction": map[string]any{
			"parts": []map[string]string{{"text": system}},
		},
		"contents": []map[string]any{
			{"role": "user", "parts": []map[string]string{{"text": user}}},
		},
		"generationConfig": map[string]any{"maxOutputTokens": 2000},
	})
	if err != nil {
		return "", err
	}
	var res struct {
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
	if err := json.Unmarshal(raw, &res); err != nil {
		return "", err
	}
	if res.Error != nil {
		return "", fmt.Errorf("gemini: %s", res.Error.Message)
	}
	var out strings.Builder
	for _, c := range res.Candidates {
		for _, part := range c.Content.Parts {
			out.WriteString(part.Text)
		}
	}
	if out.Len() == 0 {
		return "", fmt.Errorf("gemini: empty response")
	}
	return out.String(), nil
}
