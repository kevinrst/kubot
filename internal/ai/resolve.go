package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kubotdev/kubot/internal/model"
)

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
		model = "gemini-3.8-flash"
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

// SystemPrompt constrains the model to narrating: fixed plain-text shape,
// worst first, at most three findings, and no causal claims beyond what the
// evidence names. Without this the model restates the whole inventory.
func SystemPrompt() string {
	return "You explain Kubernetes health findings computed deterministically by kubot. " +
		"The findings are facts, not guesses. Explain them; never add new ones.\n\n" +
		"Write PLAIN TEXT in this exact shape — NO markdown (no #, no *, no bullets, " +
		"no bold, no tables, no code fences):\n\n" +
		"  <one sentence on overall health>\n\n" +
		"  <then the findings that matter, worst first, at most three, each as:>\n\n" +
		"  <workload and problem in one line, with the concrete numbers from the data>\n\n" +
		"  Why this is happening:\n" +
		"  <one line — ONLY when the evidence names a cause (exit code, OOMKilled, " +
		"FailedScheduling message); otherwise OMIT this block entirely. Never guess.>\n\n" +
		"  Check next:\n" +
		"  <one line: the concrete next step>\n\n" +
		"Rules:\n" +
		"- Every number comes from the findings. Never invent one.\n" +
		"- Carry every caveat into the advice.\n" +
		"- If nothing is broken, say so in one line and stop.\n" +
		"- Be brief: short lines, blank line between blocks."
}

// UserPrompt grounds a question in one inspection report.
func UserPrompt(question string, rep model.Report) string {
	raw, _ := json.MarshalIndent(rep, "", "  ")
	q := strings.TrimSpace(question)
	if q == "" {
		q = "Explain and prioritize these findings."
	}
	return q + "\n\nAnswer using ONLY the findings below. Focus on the ones relevant " +
		"to the question and ignore unrelated ones even if severe. If none are " +
		"relevant, say so and name the closest ones — do not guess.\n\n" +
		"Findings (JSON, computed by kubot — treat as facts):\n" + string(raw)
}
