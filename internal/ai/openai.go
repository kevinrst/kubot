package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// OpenAIProvider speaks /chat/completions. Any compatible endpoint
// (OpenRouter, Groq, Ollama, vLLM, …) works via BaseURL override.
type OpenAIProvider struct {
	Key     string
	ModelID string
	BaseURL string
}

func (p OpenAIProvider) Name() string     { return "openai" }
func (p OpenAIProvider) Model() string    { return p.ModelID }
func (p OpenAIProvider) Endpoint() string { return p.BaseURL }
func (p OpenAIProvider) Local() bool      { return isLocal(p.BaseURL) }

func (p OpenAIProvider) Complete(ctx context.Context, system, user string) (string, error) {
	raw, err := postJSON(ctx, strings.TrimSuffix(p.BaseURL, "/")+"/chat/completions",
		"Bearer "+p.Key, "Authorization", nil, map[string]any{
			"model": p.ModelID,
			"messages": []map[string]string{
				{"role": "system", "content": system},
				{"role": "user", "content": user},
			},
			"max_tokens": 2000,
		})
	if err != nil {
		return "", err
	}
	var res struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return "", err
	}
	if res.Error != nil {
		return "", fmt.Errorf("openai: %s", res.Error.Message)
	}
	if len(res.Choices) == 0 {
		return "", fmt.Errorf("openai: empty response")
	}
	return res.Choices[0].Message.Content, nil
}
