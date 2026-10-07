# deployment_resource_risk

Half or more of a Deployment's pods run without a memory limit.

Severity: note

## What kubot saw

Per-deployment fraction of pods missing memory limits, linked through the
ReplicaSet owner chain.

## Why it matters

One limitless container under node pressure can evict or starve its
neighbors. The risk is collective — a single bare pod is a note on the pod,
a whole bare deployment is a note on the deployment.

## Verify it yourself

```sh
kubectl get pods -l <selector> -o jsonpath='{.items[*].spec.containers[*].resources.limits}'
```

## Fix

Set memory limits across the deployment. Start with the hungriest container
(`kubot resources` shows live usage per container).

## When to ignore

Namespaces where limits are deliberately absent (dev sandboxes, batch pools):

```toml
[[ignore]]
finding = "deployment_resource_risk"
object = "dev/*"
reason = "sandbox namespace, no limits by policy"
```

## What kubot can't see

Node headroom — on an empty node the risk is theoretical. Check capacity
before treating this as urgent.
