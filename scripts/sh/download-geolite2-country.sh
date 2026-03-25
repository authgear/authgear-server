#!/bin/sh

# Download GeoLite2-Country.mmdb from MaxMind and place it at pkg/util/geoip/GeoLite2-Country.mmdb
# Requires a MaxMind license key set via MAXMIND_LICENSE_KEY environment variable.
# Usage: MAXMIND_LICENSE_KEY=<your_key> ./scripts/sh/download-geolite2-country.sh

set -eu

if [ -z "${MAXMIND_LICENSE_KEY:-}" ]; then
  echo "Error: MAXMIND_LICENSE_KEY environment variable is not set." >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
DEST="$REPO_ROOT/pkg/util/geoip/GeoLite2-Country.mmdb"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

DOWNLOAD_URL="https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-Country&license_key=${MAXMIND_LICENSE_KEY}&suffix=tar.gz"
ARCHIVE="$TMPDIR/GeoLite2-Country.tar.gz"

echo "Downloading GeoLite2-Country..."
curl -fsSL "$DOWNLOAD_URL" -o "$ARCHIVE"

echo "Extracting..."
tar -xzf "$ARCHIVE" -C "$TMPDIR"

MMDB="$(find "$TMPDIR" -name 'GeoLite2-Country.mmdb' | head -n 1)"
if [ -z "$MMDB" ]; then
  echo "Error: GeoLite2-Country.mmdb not found in archive." >&2
  exit 1
fi

cp "$MMDB" "$DEST"
echo "Done: $DEST"
