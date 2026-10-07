# ingress_no_backends

An Ingress route points at a service that doesn't exist or has no ready
endpoints.

Severity: warning

## What kubot saw

Per backend: service missing entirely (name typo, wrong namespace thinking —
backends are namespace-local) vs service present with zero ready endpoints
(backends down or unready). One finding per broken backend, so the message
names which case.

## Why it matters

That route serves 503s (or nothing). Ingress misconfiguration is invisible
until traffic arrives — nothing in the ingress *looks* broken.

## Verify it yourself

```sh
kubectl describe ingress <name>   # rules and backend refs
kubectl get endpointslices -l kubernetes.io/service-name=<svc>
```

## Fix

Missing service: fix the name (and remember backends can't cross
namespaces). No ready endpoints: fix the backing pods — readiness gates
endpoints, so this usually resolves to a pod finding.

## When to ignore

Ingresses staged before their backends exist (GitOps ordering windows):

```toml
[[ignore]]
finding = "ingress_no_backends"
object = "default/shop"
reason = "backend deploys after the ingress in the pipeline"
```

## What kubot can't see

Controller health — kubot judges ingress *config*; whether nginx/traefik
itself is alive is a workload question for the controller's own pods. TLS
termination reality too (cert validity needs the secret's contents).
