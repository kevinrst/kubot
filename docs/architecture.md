# Architecture

```
Kubernetes API (GET/LIST only, never writes)
      ↓
internal/k8s — Collector → normalized Snapshot (+ Degraded notes)
      ↓
internal/diagnose — pure Rules over the Snapshot → sorted []Finding
      ↓
internal/model — stable Report contract (versioned JSON Schema)
      ↓
+----------+----------+----------+----------+
|          |          |          |          |
CLI      JSON      MCP      AI layer
cmd/     render    mcp/     internal/ai
```

The layering is the point: `internal/k8s` performs no diagnosis, rules
perform no I/O, and every surface (text, JSON, MCP, model narration) reads
the same findings. Suppression (`.kubot.toml`) applies once, in the shared
gather path, so CLI and MCP can never disagree.

| Area | Package | Notes |
|---|---|---|
| Commands | `cmd/kubot` | one file per command + `main`/`gather`/`helpers`; no logic beyond flags |
| Collection | `internal/k8s` | snapshot, collector, metrics, client; degrades, never aborts partially |
| Engine | `internal/diagnose` | `Rule` interface, deterministic ordering, workload filtering with selector closure |
| Findings | `internal/model` | `Finding`/`Report`, `SchemaVersion`, schema generation |
| Rendering | `internal/render` | shared lipgloss palette, findings report, resources/events tables |
| MCP | `internal/mcp` | stdio JSON-RPC server, deterministic tools only |
| AI | `internal/ai` | providers (OpenAI-compatible, Anthropic, Gemini), prompt discipline |
| Config | `internal/config` | `.kubot.toml` parse + glob matching |
| Schema gen | `tools/schemagen` | single source for `schema/` |

Conventions: pure functions over the snapshot, byte-compared golden
outputs where rendering matters (table widths), drift tests for the JSON
schema and the findings catalogue, fake clients (never a live cluster) in
unit tests.
