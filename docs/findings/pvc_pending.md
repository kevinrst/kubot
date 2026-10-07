# pvc_pending

A PersistentVolumeClaim is Pending and unbound — pods using it can't start.

Severity: warning

## What kubot saw

Claim phase `Pending`, with storage class, requested size, and access modes.

## Why it matters

Unbound PVCs block pod scheduling entirely (the scheduler refuses pods with
unbound immediate claims). One stuck claim can hold a whole StatefulSet.

## Verify it yourself

```sh
kubectl get pvc
kubectl describe pvc <name>   # Events: provisioning failures
kubectl get storageclass
```

## Fix

The usual suspects: StorageClass name typo (or a class with no provisioner
on this cluster), no matching PVs for manual provisioning, or quota
exceeded.

## When to ignore

Claims waiting on purpose (pre-provisioned volumes arriving later):

```toml
[[ignore]]
finding = "pvc_pending"
object = "default/restore-*"
reason = "volume arrives with the restore job"
```

## What kubot can't see

Provisioner internals — *why* dynamic provisioning fails lives in the
external-provisioner logs, not in cluster state kubot reads.
