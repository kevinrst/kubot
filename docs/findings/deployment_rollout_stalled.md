# deployment_rollout_stalled

A rollout exceeded its progress deadline and is stuck.

Severity: warning

## What kubot saw

The `Progressing` condition with reason `ProgressDeadlineExceeded`, plus
desired/updated/ready counts showing where the rollout wedged.

## Why it matters

The new version will never finish arriving: typically the new ReplicaSet's
pods can't become ready (bad image, failing probes, crashloop). The old
version usually still serves, so this is urgency without outage — warning,
not critical.

## Verify it yourself

```sh
kubectl rollout status deployment/<name>
kubectl get replicasets -l <selector>   # the stuck new RS beside the old one
```

## Fix

Look at the *new* ReplicaSet's pods, not the deployment — that's where the
bad image or failing probe lives. Fix forward or `kubectl rollout undo`.

## When to ignore

During intentionally slow rollouts only if you raised the deadline to match —
otherwise a stalled rollout is always worth knowing about. Rarely muted:

```toml
[[ignore]]
finding = "deployment_rollout_stalled"
object = "default/big-migration"
reason = "multi-hour rollout, deadline raised accordingly"
```

## What kubot can't see

*Why* the new pods won't ready — the child pod findings (image, probe,
crash) carry that; this finding is the rollout-level rollup.
