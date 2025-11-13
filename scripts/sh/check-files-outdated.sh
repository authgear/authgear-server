#!/bin/sh

# Download or compare remote files and JSON arrays

set -eu

TMP_FILE=$(mktemp)
AUTO_UPDATE="false"

cleanup() {
  rm -f "$TMP_FILE"
}
trap cleanup EXIT INT TERM

do_compare() {
  local_file="$1"

  if [ ! -f "$local_file" ]; then
    echo "Local file missing: $file"
    exit 1
  fi

  if cmp -s "$local_file" "$TMP_FILE"; then
    echo "$local_file is up to date."
  else
    if [ "$AUTO_UPDATE" = "true" ]; then
      echo "$local_file is OUTDATED. Updating local file..."
      cp "$TMP_FILE" "$local_file"
      echo "$local_file is updated."
    else
      echo "$local_file is OUTDATED."
      exit 1
    fi
  fi
}

compare_file() {
  remote_url="$1"
  local_file="$2"

  echo "Fetching remote file: $remote_url"
  curl -sSL "$remote_url" -o "$TMP_FILE"

  do_compare "$local_file"
}

compare_json() {
  remote_url="$1"
  local_file="$2"

  echo "Fetching remote JSON array: $remote_url"
  curl -sSL "$remote_url" | jq -r '.[]' > "$TMP_FILE"

  do_compare "$local_file"
}

# Usage
usage() {
  cat <<EOF
Usage:
  $0 compare-file  <remote_url> <local_file> [--update]
  $0 compare-json  <remote_url> <local_file> [--update]

Options:
  --update     Automatically overwrite local file when differences are found.

Examples:
  $0 compare-file https://example.com/config.txt ./config.txt
  $0 compare-json https://example.com/data.json ./data.txt --update
EOF
  exit 1
}

if [ "$#" -lt 3 ]; then
  usage
fi

mode="$1"
remote_url="$2"
target_file="$3"
shift 3 || true

# Parse optional flags
for arg in "$@"; do
  case "$arg" in
    --update) AUTO_UPDATE="true" ;;
    -h|--help) usage ;;
    *) echo "Unknown option: $arg"; usage ;;
  esac
done

# Run
case "$mode" in
  compare-file)  compare_file "$remote_url" "$target_file" ;;
  compare-json)  compare_json "$remote_url" "$target_file" ;;
  *) usage ;;
esac
