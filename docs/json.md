# JSON contract

`kubot inspect --json` is the interface to build on. The terminal output is
unstable by design — parse `--json`, never the text.

- Envelope: `schema_version`, `cluster`, `status` (`ok|warning|critical`),
  `issues[]`, `checked[]`, `degraded[]`.
- Finding: `severity`, `resource` (`kind/name`), `namespace`, `reason`
  (snake_case, stable — see the [findings catalogue](findings/README.md)),
  `message`, `evidence`, `recommendation`, plus `suppressed` /
  `suppression_reason` when muted.
- Machine-checked schemas live in `schema/` (`kubot-inspect-<version>.json`),
  generated from the Go types by `go run ./tools/schemagen`. A drift test
  fails CI if the committed schema differs from a fresh generation.

## Versioning

`MAJOR` = breaking field change (removed/renamed field, changed type or
units). `MINOR` = additive fields only — an older consumer parses newer
output unchanged. History: `0.1.0` initial, `0.2.0` added `degraded`,
`0.3.0` added `suppressed` / `suppression_reason`, `1.0.0` — no field
changes, stability commitment. Parsers pin to `1.0.0`.
