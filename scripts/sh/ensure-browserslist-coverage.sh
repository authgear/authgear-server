#!/bin/sh

# This script is used to test whether .browerslistrc has enough coverage

echo 1>&2 "[INFO] Script executed from ${PWD}"

main() {
  TARGET=90
  coverage=$(npx browserslist --json --coverage | python3 -c 'import json,sys;print(json.load(sys.stdin)["coverage"]["global"])')
  expr="$coverage > $TARGET"
  result=$(echo "$expr" | bc)
  echo 1>&2 "[INFO] $expr is $result"
  if [ $result = 0 ]; then

    echo 1>&2 "[ERROR] Coverage $coverage is below $TARGET"
    exit 1
  fi
}

echo 1>&2 "[INFO] Start of script"
main
echo 1>&2 "[INFO] End of script"
