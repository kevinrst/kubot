# daemonset_unavailable

A DaemonSet has nodes without a ready pod.

Severity: critical

## What kubot saw

`numberUnavailable > 0`, with desired/ready/available/misscheduled counts.

## Why it matters

DaemonSets run per node (log agents, CNI, CSI). One bad pod spec doesn't
break one workload — it breaks the agent on *every* node at once.

## Verify it yourself

```sh
kubectl get pods -l <selector> -o wide   # which nodes are affected
kubot why <name>
```

## Fix

Check node affinity and tolerations first (the usual DaemonSet killers —
a taint the agent doesn't tolerate), then the pod logs on an affected node.

## When to ignore

During node rollouts where the agent is briefly down per node:

```toml
[[ignore]]
finding = "daemonset_unavailable"
object = "monitoring/agent"
reason = "node pool rotation this week"
```

## What kubot can't see

Whether the missing agent actually matters right now (logs may buffer and
catch up) — check the agent's own backlog.
