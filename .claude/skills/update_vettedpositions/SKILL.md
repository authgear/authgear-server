---
name: update_vettedpositions
description: Update .vettedpositions after harmless line-number moves in goanalysis output.
argument-hint: "<changed files or analysis output>"
---

Use this skill only for line-number drift in vetted positions.

## Workflow

1. Run `go run ./devtools/goanalysis ./cmd/... ./pkg/...`.
2. For each reported item, check whether it is only a line-number move with no semantic change.
3. If it is a line-number move, update `.vettedpositions` to remove the old position and add the new one.
4. If it is a genuinely new warning, stop and report it instead of updating `.vettedpositions`.
5. Re-run `go run ./devtools/goanalysis ./cmd/... ./pkg/...`.
6. Run `make sort-vettedpositions`.
7. Commit the change with the message `Update .vettedpositions for changed line numbers`.

## Notes

- Do not add new vetted positions for new issues.
- Keep the file sorted by using the repo target, not by hand.
