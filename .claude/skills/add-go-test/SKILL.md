---
name: add-go-test
description: Write or extend Go unit tests in this repo. Use when the user asks to add or update Go tests.
argument-hint: "<package or behavior>"
---

Follow this skill when adding Go tests.

## Critical: Match Existing Test Style First

**Always inspect existing `*_test.go` files in the package first.** This repo uses different testing styles in different packages:

- **Convey (BDD-style)**: Many packages use `github.com/smartystreets/goconvey`. Identified by import `. "github.com/smartystreets/goconvey/convey"` and test structure like:
  ```go
  func TestFoo(t *testing.T) {
    Convey("description", func() {
      // assertions here
      So(result, ShouldEqual, expected)
    })
  }
  ```

- **Standard `testing.T`**: Some packages use plain `*testing.T` with manual assertions.

**When in doubt, use Convey** if the package has any imports of it. Convey provides better error messages and matches the repo's BDD conventions.

## Workflow

1. **Inspect the package's `*_test.go` files first** to identify the testing style being used.
2. **Match the local test style exactly.** Do not mix styles in the same package.
3. Add the smallest test file that covers the behavior.
4. Run `go test` on the affected package. If the change crosses packages or touches shared code, run the narrowest broader test set that proves the change is safe.

## Notes

- Prefer table-driven tests when the package already uses them.
- Keep fixtures local to the test file unless they are reused elsewhere.
- If the test requires generated code or assets, rerun the generator before the final test run.
- When using Convey, use `So()` assertions for consistency with the package style, not manual `if` statements.
