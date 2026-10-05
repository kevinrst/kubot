package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kevinrst/kubot/internal/model"
)

// Resolve picks a provider from the environment. Keys are never flags.
// KUBOT_AI_PROVIDER selects explicitly; otherwise the first configured key
// wins (OpenAI, then Anthropic, then Gemini).
func Resolve() (Provider, error) {
	want := strings.ToLower(envFirst("KUBOT_AI_PROVIDER"))
	model := envFirst("KUBOT_AI_MODEL")
	base := envFirst("KUBOT_AI_BASE_URL")
	key := envFirst("KUBOT_AI_API_KEY")

	switch want {
	case "openai":
		return openAIProvider(firstNonEmpty(key, envFirst("OPENAI_API_KEY")), model, base)
	case "anthropic":
		return anthropicProvider(firstNonEmpty(key, envFirst("ANTHROPIC_API_KEY")), model, base)
	case "gemini":
		return geminiProvider(firstNonEmpty(key, envFirst("GEMINI_API_KEY", "GOOGLE_API_KEY")), model, base)
	case "":
		// auto-detect: explicit key wins, else first provider key found.
		// A lone KUBOT_AI_API_KEY (+ optional BASE_URL) means OpenAI-compatible.
		if key != "" {
			return openAIProvider(key, model, firstNonEmpty(base, "https://api.openai.com/v1"))
		}
		if k := envFirst("OPENAI_API_KEY"); k != "" {
			return openAIProvider(k, model, firstNonEmpty(base, "https://api.openai.com/v1"))
		}
		if k := envFirst("ANTHROPIC_API_KEY"); k != "" {
			return anthropicProvider(k, model, firstNonEmpty(base, "https://api.anthropic.com"))
		}
		if k := envFirst("GEMINI_API_KEY", "GOOGLE_API_KEY"); k != "" {
			return geminiProvider(k, model, firstNonEmpty(base, "https://generativelanguage.googleapis.com"))
		}
		return nil, fmt.Errorf("no model key: set OPENAI_API_KEY, ANTHROPIC_API_KEY, or GEMINI_API_KEY (keys come from env only, never flags)")
	default:
		return nil, fmt.Errorf("unknown provider %q (openai|anthropic|gemini)", want)
	}
}

func openAIProvider(key, model, base string) (Provider, error) {
	if key == "" {
		return nil, fmt.Errorf("openai needs a key: OPENAI_API_KEY or KUBOT_AI_API_KEY")
	}
	if model == "" {
		model = "gpt-4o-mini"
	}
	if base == "" {
		base = "https://api.openai.com/v1"
	}
	return OpenAIProvider{Key: key, ModelID: model, BaseURL: base}, nil
}

func anthropicProvider(key, model, base string) (Provider, error) {
	if key == "" {
		return nil, fmt.Errorf("anthropic needs a key: ANTHROPIC_API_KEY or KUBOT_AI_API_KEY")
	}
	if model == "" {
		model = "claude-sonnet-4-5"
	}
	if base == "" {
		base = "https://api.anthropic.com"
	}
	return AnthropicProvider{Key: key, ModelID: model, BaseURL: base}, nil
}

func geminiProvider(key, model, base string) (Provider, error) {
	if key == "" {
		return nil, fmt.Errorf("gemini needs a key: GEMINI_API_KEY or KUBOT_AI_API_KEY")
	}
	if model == "" {
		model = "gemini-2.5-flash"
	}
	if base == "" {
		base = "https://generativelanguage.googleapis.com"
	}
	return GeminiProvider{Key: key, ModelID: model, BaseURL: base}, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// SystemPrompt instructs the model to narrate findings, never invent them.
func SystemPrompt() string {
	return "You explain Kubernetes health findings computed deterministically by kubot. " +
		"Treat the provided findings as facts: never invent problems, causes, or numbers not present. " +
		"Carry every caveat into your advice. Prioritize by severity (critical first). " +
		"Keep it short: what is wrong, likely cause, what to check next."
}

// UserPrompt grounds a question in one inspection report.
func UserPrompt(question string, rep model.Report) string {
	raw, _ := json.MarshalIndent(rep, "", "  ")
	q := strings.TrimSpace(question)
	if q == "" {
		q = "Explain and prioritize these findings."
	}
	return q + "\n\nFindings (JSON, computed by kubot — treat as facts):\n" + string(raw)
}
