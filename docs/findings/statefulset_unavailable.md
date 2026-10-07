# statefulset_unavailable

A StatefulSet has replicas that aren't ready.

Severity: critical

## What kubot saw

Ready replicas below desired (`spec.replicas`, default 1 — StatefulSets
don't report unavailable directly, so kubot computes it), with
current/updated counts.

## Why it matters

Stateful workloads (databases, queues) are usually singletons or quorums —
one missing replica can mean no writes, not just less capacity.

## Verify it yourself

```sh
kubectl get pods -l <selector>   # ordered pods: -0, -1, -2…
kubot why <name>
```

## Fix

Stateful pods block in order: check the lowest ordinal first. Common
blockers are storage (unbound PVC — see `pvc_pending`) and identity/startup
ordering, not just bad images.

## When to ignore

Same as deployments — only during a handled incident window:

```toml
[[ignore]]
finding = "statefulset_unavailable"
object = "db/*"
reason = "planned failover test"
```

## What kubot can't see

Quorum state inside the app (is the database actually writable?) — that
needs the app's own health endpoint.
