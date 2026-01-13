First, run `make ensure-important-modules-up-to-date` to check what modules are outdated.

Then, in the following directories, update the module:
  - ./
  - ./custombuild
  - ./e2e

Use these commands to update:
```
go get -u {MODULE}
go mod tidy
```
