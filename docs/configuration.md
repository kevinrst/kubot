# Configuration

One file, one job: muting tolerated noise. `.kubot.toml` lives beside your
manifests and is meant to be committed.

```toml
[[ignore]]
finding = "deployment_resource_risk"
reason = "limits rollout scheduled next quarter"

[[ignore]]
finding = "pod_missing_resources"
object = "default/batch-*"
reason = "short-lived jobs, limits add nothing"
```

Discovery order: `--config path`, then `$KUBOT_CONFIG`, then `./.kubot.toml`.
No file, no behavior change.

## Contract

- `finding` (required): the exact `reason` id from `--json`.
- `object` (optional): glob matched against `namespace/resource`
  (`kube-system/*`), the bare resource (`pod/coredns-*`), and the
  `namespace/name` form (`default/limitless-*`, kind prefix optional).
  A trailing `/*` matches the whole subtree; omit `object` to mute the
  finding everywhere.
- `reason` (required): why it's tolerated. Rejected if empty — the config
  loader fails the run rather than accept a silent mute.

Suppressed findings stay visible (marked with the reason) and stay in
`--json` (`suppressed: true`), but never move exit codes, status, or the
health score. Credential-shaped keys (`password`, `token`, `url`, …) are
refused outright — secrets don't belong in committed files.
