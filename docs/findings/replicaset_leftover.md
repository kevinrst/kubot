# replicaset_leftover

Old ReplicaSets still run pods beside the newest one.

Severity: note

## What kubot saw

Two or more ReplicaSets owned by one Deployment with replicas above zero —
the current (newest) plus old ones that should have scaled down, named in
evidence.

## Why it matters

Usually a stuck rollout's shadow: the old version still serving means the
new one never took over. Occasionally manual scaling someone forgot to
revert.

## Verify it yourself

```sh
kubectl get replicasets -l <selector>
```

## Fix

If a rollout is stuck, fix that (`deployment_rollout_stalled`). If the
rollout finished, scale the leftover down or let the controller finish.

## When to ignore

Blue-green style setups that intentionally keep the old set warm:

```toml
[[ignore]]
finding = "replicaset_leftover"
object = "default/web"
reason = "blue-green keeps previous set warm"
```

## What kubot can't see

Intent — kubot can't distinguish "stuck" from "deliberately warm". The
stalled-rollout finding firing alongside is the tell for stuck.
