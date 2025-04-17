#!/bin/sh

# List out all direct (i.e. not Indirect) modules that have update.
# Also list out the following listed modules that have update.
go list -u -m -json all | jq --raw-output --argjson packages '
[
  "golang.org/x/crypto",
  "golang.org/x/exp",
  "golang.org/x/image",
  "golang.org/x/mod",
  "golang.org/x/net",
  "golang.org/x/oauth2",
  "golang.org/x/sync",
  "golang.org/x/sys",
  "golang.org/x/telemetry",
  "golang.org/x/term",
  "golang.org/x/text",
  "golang.org/x/time",
  "golang.org/x/tools",
  "golang.org/x/vuln"
]
' '
if .Update != null and ((.Indirect | not) or ([.Path] | inside($packages))) then
  "\(.Path) \(.Version) [\(.Update.Version)]"
else
  null
end | values
'
