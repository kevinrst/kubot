# ingress_tls_secret_missing

An Ingress TLS block references a secret that doesn't exist.

Severity: warning

## What kubot saw

`spec.tls[]` entries whose `secretName` has no Secret in the ingress
namespace, with the hosts listed as evidence.

## Why it matters

TLS terminates nowhere: HTTPS either fails or falls back to a default cert
(depending on controller), and the failure mode is a browser error, not a
cluster event.

## Verify it yourself

```sh
kubectl get secret -n <ns>   # is it there, spelled right, same namespace?
```

## Fix

Create it (`kubectl create secret tls <name> --cert=... --key=...`) or fix
the name. Secrets are namespace-local — the most common cause is creating it
in the wrong namespace.

## When to ignore

TLS terminated elsewhere (external LB, CDN in front) with the block left as
documentation:

```toml
[[ignore]]
finding = "ingress_tls_secret_missing"
object = "default/shop"
reason = "TLS terminates at the CDN"
```

## What kubot can't see

Cert *validity* — expiry, hostname match, chain. kubot checks existence,
not contents. Check those with openssl against the live endpoint.
