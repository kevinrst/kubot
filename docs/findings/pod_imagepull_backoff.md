# pod_imagepull_backoff

A container (or init container) can't pull its image.

Severity: critical

## What kubot saw

Container waiting with reason `ImagePullBackOff` or `ErrImagePull`, plus the
image reference and the registry's error message.

## Why it matters

The pod will never start. Nothing downstream (probes, readiness, traffic)
matters until the image resolves.

## Verify it yourself

```sh
kubectl describe pod <pod>   # Events: the exact registry error
```

## Fix

Three usual causes, in order: typo'd tag or repo name, image never pushed,
missing/wrong `imagePullSecrets` for a private registry. The event message
names which one.

## When to ignore

Practically never — an unpullable image is always broken. The only mute is a
registry you know is temporarily down:

```toml
[[ignore]]
finding = "pod_imagepull_backoff"
object = "default/*"
reason = "registry outage 2026-10-07, revert after"
```

## What kubot can't see

Registry credentials validity beyond what kubelet reports, and whether the
tag exists upstream — check the registry directly.
