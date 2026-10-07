# pod_mount_failure

A pod can't mount a volume (PVC, secret, configMap — anything kubelet mounts).

Severity: warning

## What kubot saw

Repeated `FailedMount` warning events (3+ aggregated, or any count on a pod
older than 15 minutes — fresh event objects restart the counter hourly), with
the kubelet's message.

## Why it matters

The container may run, but without its volume it's wrong: missing config,
missing credentials, missing data. Mount failures are silent-ish — the pod
can sit in `ContainerCreating` indefinitely.

## Verify it yourself

```sh
kubectl describe pod <pod>   # FailedMount events with the exact error
kubectl get pvc,secret,configmap -n <ns>   # does the source exist?
```

## Fix

Match the message: missing secret/configMap (create it, check the name and
namespace), unbound PVC (see `pvc_pending`), or kubelet unable to reach the
volume backend.

## When to ignore

Volumes that arrive late by design (secrets injected by an operator that
hasn't run yet):

```toml
[[ignore]]
finding = "pod_mount_failure"
object = "default/vault-wait-*"
reason = "secret appears when the injector runs"
```

## What kubot can't see

CSI driver internals — backend-side attach failures need the driver's logs.
