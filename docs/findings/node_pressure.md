# node_pressure

A node reports memory, disk, or PID pressure.

Severity: warning

## What kubot saw

Node conditions `MemoryPressure`, `DiskPressure`, or `PIDPressure` true,
listed in evidence.

## Why it matters

Pressure precedes eviction: the kubelet starts killing pods (usually the
limitless, BestEffort ones first — see `pod_missing_resources`) to relieve
the node. A pressure warning is advance notice of seemingly random pod deaths.

## Verify it yourself

```sh
kubectl describe node <name>   # Conditions and what trips them
kubectl top node <name>
```

## Fix

Add capacity, drain the node, or find what's consuming the pressured
resource. For disk pressure: logs and container overlays are the usual
suspects. For PID pressure: something is forking out of control.

## When to ignore

Single-node dev clusters perpetually near limits:

```toml
[[ignore]]
finding = "node_pressure"
reason = "kind node runs hot, nothing to do"
```

## What kubot can't see

*What* consumes the resource — node conditions are boolean, not accounting.
`kubectl top` and the kubelet metrics carry that.
