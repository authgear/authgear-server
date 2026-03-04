1. Run `go run ./devtools/goanalysis ./cmd/... ./pkg/...` to get the list of errors.
2. For each error, determine whether it is a simple line number change (i.e. the same identifier/rule moved to a different line, with no semantic change):
  - If it IS a simple line number change: update .vettedpositions to remove the old position and add the new one.
  - If it is NOT a simple line number change (e.g. a genuinely new unvetted usage, a new file, a different identifier): do NOT update .vettedpositions. Instead, notify the user and stop.
3. Run `go run ./devtools/goanalysis ./cmd/... ./pkg/...` again to ensure your changes are ok.
4. Run `make sort-vettedpositions` to sort the lines.
5. Make a commit with the message: "Update .vettedpositions for changed line numbers"
