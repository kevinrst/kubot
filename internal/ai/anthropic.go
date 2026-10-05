package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type AnthropicProvider struct {
	Key     string
	ModelID string
	BaseURL string
}

func (p AnthropicProvider) Name() string     { return "anthropic" }
func (p AnthropicProvider) Model() string    { return p.ModelID }
func (p AnthropicProvider) Endpoint() string { return p.BaseURL }
func (p AnthropicProvider) Local() bool      { return isLocal(p.BaseURL) }

func (p AnthropicProvider) Complete(ctx context.Context, system, user string) (string, error) {
	raw, err := postJSON(ctx, strings.TrimSuffix(p.BaseURL, "/")+"/v1/messages",
		p.Key, "x-api-key",
		map[string]string{"anthropic-version": "2023-06-01"},
		map[string]any{
			"model":      p.ModelID,
			"max_tokens": 2000,
			"system":     system,
			"messages": []map[string]string{
				{"role": "user", "content": user},
			},
		})
	if err != nil {
		return "", err
	}
	var res struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return "", err
	}
	if res.Error != nil {
		return "", fmt.Errorf("anthropic: %s", res.Error.Message)
	}
	var out strings.Builder
	for _, b := range res.Content {
		if b.Type == "text" {
			out.WriteString(b.Text)
		}
	}
	if out.Len() == 0 {
		return "", fmt.Errorf("anthropic: empty response")
	}
	return out.String(), nil
}
