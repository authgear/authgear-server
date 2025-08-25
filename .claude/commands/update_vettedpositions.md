1. Run `go run ./devtools/goanalysis ./cmd/... ./pkg/...` to get the list of errors.
2. The current code implementation is correct, therefore, you should update .vettedpositions to suppress the errors:
  - You should remove unused vetted positions from the file
  - You should update the file to suppress new unvetted usage errors
3. Run `go run ./devtools/goanalysis ./cmd/... ./pkg/...` again to ensure your changes are ok.
4. Finally, run `make sort-vettedpositions` to sort the lines.
