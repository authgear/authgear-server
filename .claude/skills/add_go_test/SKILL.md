---
name: add_go_test
description: Write or extend Go unit tests in this repo. Use when the user asks to add or update Go tests.
argument-hint: "<package or behavior>"
---

Follow this skill when adding Go tests.

## Workflow

1. Inspect the package and its existing `*_test.go` files first.
2. Match the local test style. If the package already uses `github.com/smartystreets/goconvey`, keep using it; otherwise follow the surrounding pattern.
3. Add the smallest test file that covers the behavior.
4. Run `go test` on the affected package. If the change crosses packages or touches shared code, run the narrowest broader test set that proves the change is safe.

## Notes

- Prefer table-driven tests when the package already uses them.
- Keep fixtures local to the test file unless they are reused elsewhere.
- If the test requires generated code or assets, rerun the generator before the final test run.
