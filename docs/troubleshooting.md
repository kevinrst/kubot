# Troubleshooting

## `cannot reach cluster: connection refused`

kubot can't dial the API server — and says so loudly (exit 3) instead of
reporting an empty "healthy". Check the context points somewhere alive
(`kubectl cluster-info`), Docker Desktop is running (for local clusters),
and no VPN/firewall eats the port. On Windows, Hyper-V can claim the port
range of a kind cluster's published port after a restart; a socat forwarder
on a free port plus a second context is the workaround.

## Findings appear and disappear between runs

Two honest causes, both by design. Kubernetes events expire after ~1 hour,
so count-based rules (probes, mounts) rebuild from 1 — sustained problems
re-fire within minutes; one-off blips correctly stay silent the second
time. And the crash rule only counts crashes from the last 15 minutes, so a
restart storm (e.g. a Docker outage) stops paging once it's over.

## `pod_near_limit` never fires / `degraded` mentions metrics

The rule needs metrics-server. Most managed clusters ship it; `kind` does
not. Everything else works without it.

## `usage_ratio` looks wrong

When only limits are set, Kubernetes defaults requests = limits — so a pod
with "no requests" may still show a request value. The table and findings
report what the API returns, including defaults.

## No color / broken box-drawing

`--no-color`, `NO_COLOR`, and non-TTY output disable color. Crooked table
borders are a terminal font without box-drawing glyphs — switch to
JetBrains Mono, Cascadia Code, or Fira Code.

## Stale model ids

Model ids retire (`gemini-2.5-flash` 404'd on us). The error names the
`KUBOT_AI_MODEL` override — set it to a live id and retry.
