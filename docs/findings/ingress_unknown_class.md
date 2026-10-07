# ingress_unknown_class

An Ingress names a class with no installed IngressClass, or sets none where
no default exists.

Severity: warning

## What kubot saw

Explicit `ingressClassName` with no matching IngressClass object, or empty
class with no default-class annotation anywhere in the cluster.

## Why it matters

No matching controller picks the ingress up — it's config that nothing
reconciles. Silently dead.

## Verify it yourself

```sh
kubectl get ingressclass
```

## Fix

Install the controller for the class, fix the class name typo, or annotate
one class as default (`ingressclass.kubernetes.io/is-default-class=true`).

## When to ignore

Clusters mid-migration between ingress controllers:

```toml
[[ignore]]
finding = "ingress_unknown_class"
object = "default/*"
reason = "migrating nginx to traefik this week"
```

## What kubot can't see

Whether the *controller* behind an existing class is actually running —
class existence is not controller health. Check the controller pods.
