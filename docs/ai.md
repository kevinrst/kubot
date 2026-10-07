# AI layer

`ask` and `explain` put a plain-language reading on top of the same
deterministic findings. The model narrates; it never diagnoses — findings
are computed in Go either way, and the deterministic report stands alone if
the model fails or no key is set.

```sh
export GEMINI_API_KEY=…        # or OPENAI_API_KEY, or ANTHROPIC_API_KEY
kubot ask "which of these can wait until morning?"
kubot explain payments-api
```

## Providers and overrides

| Variable | Purpose |
|---|---|
| `OPENAI_API_KEY` / `ANTHROPIC_API_KEY` / `GEMINI_API_KEY` | enables the layer (auto-detected in that order) |
| `KUBOT_AI_PROVIDER` | `openai`, `anthropic`, or `gemini` — pin one explicitly |
| `KUBOT_AI_MODEL` | model id override (defaults rot — this is the escape hatch) |
| `KUBOT_AI_BASE_URL` | any OpenAI-compatible endpoint (Ollama, vLLM, OpenRouter, …) |
| `KUBOT_AI_API_KEY` | key override for the selected provider |

Keys come from env only, never flags.

## Disclosure rules

Before anything leaves the machine, kubot names the provider and model and
asks (`--yes` skips in scripts; a non-terminal without `--yes` refuses).
Local endpoints skip confirmation — nothing leaves the box. Model ids
retire without warning: a 404 names the `KUBOT_AI_MODEL` override in the
error.

## Prompt discipline

The system prompt pins the model down: findings are facts, plain-text
shape, worst first, at most three, no markdown, causal claims only from
named evidence (otherwise the block is omitted), every number from the
data. The `ask` question additionally scopes the answer to relevant
findings — unrelated ones are ignored even if severe. Output is labeled
with provider/model and "verify before acting", in magenta reserved for
model text.
