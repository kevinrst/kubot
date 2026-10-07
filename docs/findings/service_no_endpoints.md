# service_no_endpoints

A Service selects pods but has no ready backends.

Severity: warning

## What kubot saw

Zero ready addresses in the EndpointSlices, plus how many pods the selector
actually matches — which splits the diagnosis in two: selector matches
nothing (wrong labels) vs matches pods that aren't ready (backends down).

## Why it matters

Traffic to this service goes nowhere. ClusterIP, DNS, and ingress all
resolve; nothing answers.

## Verify it yourself

```sh
kubectl get endpointslices -l kubernetes.io/service-name=<svc>
kubectl get pods --show-labels | grep <selector>
```

## Fix

No matched pods: fix the selector or the pod labels (`app: web` vs
`app: webapp` is the classic). Matched-but-unready pods: fix the pods —
readiness gates endpoints.

## When to ignore

For headless/external services kubot already skips those itself. For a
service scaled to zero on purpose:

```toml
[[ignore]]
finding = "service_no_endpoints"
object = "default/night-worker"
reason = "scaled to zero overnight by cron"
```

## What kubot can't see

DNS and kube-proxy health — if endpoints exist but traffic still fails,
the problem is below kubot's read-only layer (CNI, kube-proxy, DNS).
