# job_failed

A Job recorded failures with no successes and nothing still running.

Severity: warning

## What kubot saw

`failed > 0` with `active == 0` and `succeeded == 0`, plus the backoff
limit. Still-running or partially succeeding jobs stay silent — those are
pod problems, not job verdicts.

## Why it matters

The work never completed, and Kubernetes has stopped retrying (or will
stop when the backoff budget runs out). Failed jobs sit invisible unless
something looks at them — which is the point of this rule.

## Verify it yourself

```sh
kubectl get jobs
kubectl logs job/<name>   # the failed attempt's output
```

## Fix

Read the failed pod logs, fix the cause, and either delete the job (a fresh
run starts clean) or raise `backoffLimit` if the failures were transient
(node evaporations, registry blips).

## When to ignore

Jobs that fail as part of their design (probes that intentionally fail
closed, chaos experiments):

```toml
[[ignore]]
finding = "job_failed"
object = "default/chaos-*"
reason = "chaos jobs fail by design"
```

## What kubot can't see

Whether the work *mattered* — a failed backup job and a failed report job
look identical here. You know which one pages.
