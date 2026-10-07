# pod_missing_resources

Containers without a memory limit, or without CPU/memory requests.

Severity: note

## What kubot saw

Per container, which of the three settings is absent. Note, not warning:
missing settings are risk, not outage — they never move exit codes.

## Why it matters

No memory limit means one hungry container can OOMKill its neighbors when
the node fills. No requests means the scheduler places blindly and QoS
drops to BestEffort. Kubernetes even defaults a missing memory *request*
to the limit, which surprises people.

## Verify it yourself

```sh
kubectl get pod <pod> -o jsonpath='{.spec.containers[*].resources}'
```

## Fix

Set all three per container. Limits cap blast radius, requests make
scheduling honest. (Static pods and `kube-system` are skipped — you can't
meaningfully patch those.)

## When to ignore

Batch jobs intentionally bare, or namespaces where you accept BestEffort:

```toml
[[ignore]]
finding = "pod_missing_resources"
object = "default/batch-*"
reason = "short-lived jobs, limits add nothing"
```

## What kubot can't see

Whether the missing values would actually matter — a sleep-3600 pod needs
nothing, and kubot flags it anyway. Notes are suggestions, treat them so.
