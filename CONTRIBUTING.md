# Contributing to kubot

Thanks for helping. kubot is a Kubernetes diagnostics CLI; a few
invariants are load-bearing, so please read the short list below before a PR.

## Build and test

```sh
go build ./cmd/kubot
gofmt -l cmd internal tools   # must print nothing
go vet ./...
staticcheck ./...            # if installed; CI-grade, catches dead code
go test ./...                # fake clients throughout — no cluster needed
```

New rules need fixtures in both directions: a case that fires and a case
that stays silent (see `rules_test.go`). If you can, also verify live
against a local cluster before pushing:

```sh
kind create cluster
kubectl apply -f <broken-workload>  # CrashLoop, bad image, OOM, pending…
kubot inspect --context kind-kind
```

## The invariants (please don't break these)

- **Read-only, always.** The collector only GETs/LISTs via client-go. No
  create/update/delete/patch anywhere in the path — kubot diagnoses
  clusters, never touches them.
- **Findings are deterministic.** Every finding is a pure function over a
  `Snapshot` (`internal/diagnose`). The AI layer narrates findings; it never
  generates them, and the report stands alone when no key is set.
- **Criticals mean broken right now.** A rule that fires on stale history
  (old restarts, expired events, healthy-again pods) is a bug, not a
  finding — see the recency bars (`crashRecency`, `minEventRepeats`) and
  the Ready-pod guards. Notes never move exit codes.
- **No credentials in output.** Findings carry cluster-state fields only.
  Never log kubeconfig contents, tokens, or keys; keys come from env, never
  flags.
- **`--json` is additive.** Add fields with `omitempty`, or bump
  `model.SchemaVersion` and regenerate (`go run ./tools/schemagen` — the
  drift test enforces it).
- **Every rule has a docs page.** A new finding needs
  `docs/findings/<reason>.md` (what it saw, how to verify, when to ignore);
  `TestCatalogueCoversRules` enforces the pairing.
- **Suppressed stays visible.** Mutes mark findings, never delete them, and
  never move exit codes or the score.

## Releasing

Tag `v*` → `.github/workflows/release.yml` → goreleaser builds the six
`linux|darwin|windows × amd64|arm64` archives → cosign signs checksums →
GitHub Release. Verify with `install.sh` + `KUBOT_REQUIRE_SIGNATURE=1`
before announcing.

## Commits

Conventional-commit prefixes (`feat:`, `fix:`, `docs:`, `test:`, `chore:`,
`ci:`) keep the changelog readable. Keep changes focused; commit each
logical unit.
