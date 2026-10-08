# kubot

A CLI that diagnoses Kubernetes instead of making you dig through it.

`kubectl get pods` says CrashLoopBackOff. kubot tells you the container OOMKilled on a 64Mi limit, 44 restarts, exit 137 — and what to check next.

Read-only. Deterministic. For humans and AI agents. [Apache-2.0](LICENSE).

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/kubotdev/kubot/main/install.sh | sh
```

Or build from source (needs Go):

```sh
go install github.com/kubotdev/kubot/cmd/kubot@latest
```

The script verifies the sha256 checksum always, and the cosign signature
(keyless, GitHub Actions OIDC) when `cosign` is on PATH —
`KUBOT_REQUIRE_SIGNATURE=1` makes the signature mandatory.

Needs a kubeconfig that can read the cluster (`~/.kube/config`, `$KUBECONFIG`, `--kubeconfig`, or in-cluster). It only GETs/LISTs. It never writes anything.

## Use

```sh
kubot inspect                      # whole cluster, all namespaces
kubot inspect payments-api         # one workload
kubot why payments-api             # why is it unhealthy
kubot events                       # warning events by count
kubot resources                    # per-container usage table + resource problems
kubot networking                   # service + probe problems
kubot check --fail-on=critical     # CI gate, exits non-zero when sick
kubot inspect --json               # machine-readable report
kubot ask "why is it slow?"        # AI reading of the same findings
kubot explain                      # report plus an AI reading of it
kubot mcp                          # serve findings to AI agents over MCP
```

Global flags: `--kubeconfig`, `--context`, `-n/--namespace`, `--timeout`, `--no-color`.

Example:

```
$ kubot why payments-api

  pod/payments-api-68ffdf658-m2wq4 (default)
    Container "api" was OOMKilled (exit 137)

    Evidence:
      · container: api
      · exit_code: 137
      · memory_limit: 64Mi
      · restart_count: 44

    Recommendation:
      Increase the container memory limit or investigate memory usage.
```

Misspell a workload and it helps instead of shrugging:

```
$ kubot why payment-api
No problems detected for "payment-api".
Did you mean: payments-api?
```

## What it checks

- CrashLoopBackOff, including the backoff window between restarts
- OOMKilled, with the memory limit as evidence
- ImagePullBackOff / ErrImagePull, with the registry error
- Pending pods that the scheduler can't place, with requests vs capacity
- Failing readiness/liveness/startup probes (sustained only, Ready pods stay silent)
- Deployments with unavailable replicas
- Services whose selector matches no pods
- Containers missing memory limits or requests
- Suspiciously low memory limits (under 32Mi)
- Containers using 80%+ of their memory limit (needs metrics-server, degrades without it)
- Deployments where half or more pods run without a memory limit
- Rollouts past their progress deadline
- Old ReplicaSets still running beside the newest
- Pending (unbound) PVCs
- Pods failing to mount volumes
- Nodes under memory/disk/PID pressure
- StatefulSets and DaemonSets with unavailable pods
- Ingresses pointing at missing or endpoint-less services
- Ingresses with unknown or missing ingress class
- Ingresses referencing nonexistent TLS secrets
- Failed Jobs with no successes
- CronJobs whose latest run failed with no success since

Twenty-three checks done well beats thirty done badly. More coming.

## JSON contract

`kubot inspect --json` emits a versioned report (`schema_version`, currently `1.0.0`, schema in `schema/`):

```json
{
  "schema_version": "1.0.0",
  "status": "critical",
  "issues": [
    {
      "severity": "critical",
      "resource": "pod/payments-api-68ffdf658-m2wq4",
      "reason": "pod_oom_killed",
      "message": "Container \"api\" was OOMKilled (exit 137)",
      "evidence": {"memory_limit": "64Mi", "restart_count": 44, "exit_code": 137},
      "recommendation": "Increase the container memory limit or investigate memory usage."
    }
  ]
}
```

Parse `--json`, not the terminal output. The human report is unstable by design; the JSON is versioned (breaking changes bump MAJOR, additions bump MINOR).

## Exit codes

`0` clean · `1` warning · `2` critical · `3` connection failure · `64` bad flags. `inspect` exits 0 by default; gate with `--fail-on` or use `check` in CI.

## Suppression

Tolerated noise gets muted, never hidden. A `.kubot.toml` beside the repo (or `--config`, or `$KUBOT_CONFIG`):

```toml
[[ignore]]
finding = "deployment_resource_risk"
reason = "limits rollout is scheduled next quarter"

[[ignore]]
finding = "pod_missing_resources"
object = "default/limitless-*"
reason = "batch job, intentionally bare"
```

Muted findings stay visible (marked with the reason) and in `--json`, but never move the exit code or health score. Rules without a reason are rejected — write why. Credential-shaped keys are refused outright.

Every finding has a reference page under [`docs/findings/`](docs/findings/README.md) — what it observed, how to verify it yourself with read-only kubectl, and a pasteable ignore block.

## AI layer (optional)

`ask` and `explain` put a plain-language reading on top of the same deterministic findings. The model narrates; it never diagnoses. Before sending anything, kubot names the provider and model and asks — local endpoints skip the prompt, `--yes` skips it in scripts.

```sh
export OPENAI_API_KEY=sk-…        # or ANTHROPIC_API_KEY, or GEMINI_API_KEY
kubot ask "what needs attention?"
kubot explain payments-api
```

| Variable | Purpose |
|---|---|
| `OPENAI_API_KEY` / `ANTHROPIC_API_KEY` / `GEMINI_API_KEY` | Enables `ask` / `explain` (auto-detected in that order) |
| `KUBOT_AI_PROVIDER` | `openai`, `anthropic`, or `gemini` — pick one explicitly |
| `KUBOT_AI_MODEL` | Model override (defaults work, override for newer ones) |
| `KUBOT_AI_BASE_URL` | OpenAI-compatible endpoint (Ollama, vLLM, OpenRouter, …) |
| `KUBOT_AI_API_KEY` | Key override for whichever provider is selected, plus optional base URL for local models |

## MCP

`kubot mcp` speaks the Model Context Protocol on stdio: `inspect` and `why` tools, a `diagnose` prompt, a `kubot://schema` resource. The agent narrates; kubot provides the facts. Point it at a cluster:
```json
{
  "mcpServers": {
    "kubot": {
      "command": ["kubot", "--context", "my-cluster", "mcp"]
    }
  }
}
```

## Develop

```sh
go build ./...
go test ./...
go run ./tools/schemagen   # regenerate schema/ from the Go types after changing them
```

Full docs live in [`docs/`](docs/README.md) — install, commands, config, JSON, MCP, AI, findings, troubleshooting, architecture.

Contributions welcome — read [`CONTRIBUTING.md`](CONTRIBUTING.md) first.

Status: early. Expect sharp edges.

## License

[Apache-2.0](LICENSE).
