# MCP

`kubot mcp` speaks the Model Context Protocol on stdio: deterministic tools
for an AI agent, with the model doing the explaining. The agent narrates;
kubot provides the facts.

- Tools: `inspect` (cluster or workload findings as JSON), `why` (requires
  `workload`) — both accept `namespace`.
- Prompt: `diagnose` (one-click "inspect and prioritize", optional workload).
- Resource: `kubot://schema` (the versioned inspect JSON Schema).

Point it at a cluster via flags (stdio has no per-call connection):

```json
{
  "mcpServers": {
    "kubot": {
      "command": "kubot",
      "args": ["--context", "my-cluster", "mcp"]
    }
  }
}

For opencode (`opencode.json`):

```json
{
  "mcp": {
    "kubot": {
      "type": "local",
      "command": ["kubot", "--context", "my-cluster", "mcp"],
      "enabled": true
    }
  }
}
```

Verified end to end with an independent JSON-RPC client: handshake,
notifications, tool/prompt/resource listing, live `why` diagnosis, and
unknown-tool rejection. MCP exposes deterministic tools only — the `ask`
model-calling layer is deliberately not servable to agents.
