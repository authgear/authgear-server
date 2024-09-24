#!/bin/sh
# This script is used to test whether the ./.browerslistrc has enough coverage

coverage=$(npx browserslist --json --coverage | python3 -c 'import json,sys;print(json.load(sys.stdin)["coverage"]["global"])')
TARGET=90
expr="$coverage > $TARGET"
result=$(echo "$expr" | bc)

echo 1>&2 "$expr: $result"

if [ $result = 0 ]; then
  exit 1
fi
