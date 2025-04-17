#!/bin/sh

output="$(go list -u -m -json all | jq --raw-output --argjson packages '
[
  "github.com/nyaruka/phonenumbers"
]
' '
if .Update != null and ([.Path] | inside($packages)) then
  "\(.Path) \(.Version) [\(.Update.Version)]"
else
  null
end | values
')"

if [ -n "$output" ]; then
  printf "%s\n" "$output"
  exit 1
fi
