# deployment_unavailable

A Deployment has replicas that aren't ready.

Severity: critical

## What kubot saw

`unavailableReplicas > 0`, or the `Available` condition not `True`, with
desired/ready/available counts and how many owned pods are unhealthy.

## Why it matters

This is the workload-level rollup: something below is broken and capacity is
reduced. The finding points at the deployment; the child pod findings say why.

## Verify it yourself

```sh
kubectl rollout status deployment/<name>
kubot why <name>   # the owned failing pods, linked automatically
```

## Fix

Fix the pods (`kubot why` shows them). If the pods look fine, check the
rollout itself — a stuck one gets its own finding (`deployment_rollout_stalled`).

## When to ignore

During a known-bad deploy window you're already handling:

```toml
[[ignore]]
finding = "deployment_unavailable"
object = "default/canary"
reason = "canary intentionally half-broken during experiment"
```

## What kubot can't see

Whether the reduced capacity actually hurts users — that depends on load and
PDBs, not replica counts.
