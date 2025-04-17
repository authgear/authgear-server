#!/bin/sh

output="$(go list -u -m -json all | jq --raw-output --argjson packages '
[
  "github.com/nyaruka/phonenumbers",
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
[
  if [.Path] | inside($packages) then
    if .Update != null then
      "\(.Path) \(.Version) [\(.Update.Version)]"
    else
      null
    end
  else
    null
  end
] | map(select(. != null)) | .[]
')"

if [ -n "$output" ]; then
  printf "%s\n" "$output"
  exit 1
fi
