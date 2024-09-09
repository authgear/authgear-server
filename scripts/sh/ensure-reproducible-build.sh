#!/bin/sh

# This script is used to test whether the built bundles are having the same hash over several builds

echo 1>&2 "[INFO] Script executed from ${PWD}"

# Arguments for the script:
# -n, the number of iterations
# -p, the path to the manifest json file
# Usage: ./ensure-reproducible-build.sh -n 5 -p manifest.json
while getopts "n:p:" option; do
  case "${option}" in
  n) num_iterations=${OPTARG} ;;
  p) manifest_location=${OPTARG} ;;
  *)
    echo 1>&2 '[ERROR] Unknown argument'
    exit 1
    ;;
  esac
done

if [ -z "$num_iterations" ] || [ -z "$manifest_location" ]; then
  echo 1>&2 "[ERROR] Missing -n or -p argument"
  exit 1
fi

if [ "$num_iterations" -lt 2 ]; then
  echo 1>&2 "[ERROR] The build number must be greater than 1"
  exit 1
fi

TEMP_DIR="$(mktemp -d)"
echo 1>&2 "[INFO] TEMP_DIR is ${TEMP_DIR}"

build_bundles() {
  for i in $(seq 1 "$num_iterations"); do
    echo 1>&2 "[INFO] Build no.$i"

    # Clear vite cache before building bundles
    rm -rf node_modules/.vite >/dev/null 2>&1
    npm run build >/dev/null 2>&1

    python3 -m json.tool --sort-keys --no-ensure-ascii --indent 2 <"$manifest_location" >"$TEMP_DIR/$i.json"
  done
}

compare_bundles() {
  echo 1>&2 "[INFO] Comparing $i builds..."

  for j in $(seq 2 "$num_iterations"); do
    diff -u "$TEMP_DIR/1.json" "$TEMP_DIR/$j.json"
    # Used to check the error code
    exit_code=$?
    if [ $exit_code -ne 0 ]; then
      echo 1>&2 "[ERROR] Build hashes are not identical"
      exit 1
    fi
  done

  echo 1>&2 "[INFO] Everything works fine, Bye!"
}

main() {
  build_bundles
  compare_bundles
}

echo 1>&2 "[INFO] Start of script"
main
echo 1>&2 "[INFO] End of script"
