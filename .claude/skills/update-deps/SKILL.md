---
name: update-deps
description: Autonomously update Go version, dependencies, CI actions, linter, and modernize code to latest Go idioms
argument-hint: "[scope]"
context: fork
agent: nepenthe-updater
allowed-tools: Bash, Read, Edit, Write, Grep, Glob, WebSearch, WebFetch
---

# Update Dependencies: $ARGUMENTS

Run the full update procedure as described in your agent instructions.

## Scope

The optional `$ARGUMENTS` parameter controls what to update. If empty or "all", run the full procedure. Otherwise, run only the matching phases:

- `go` — Update Go version only (Phase 2, step 1)
- `deps` — Update Go module dependencies only (Phase 3)
- `ci` — Update CI workflow action versions only (Phase 4)
- `lint` — Update golangci-lint version only (Phase 2, step 2)
- `modernize` — Code modernization only, no version changes (Phase 5)
- `all` or empty — Full update (all phases)

Regardless of scope, always run Phase 1 (research) for the relevant components and Phase 6 (verify) at the end. Always produce the Phase 7 summary report.
