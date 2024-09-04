#!/bin/sh

# This script is used to test whether the built bundles are having the same hash over several builds

N=5
MANIFEST_LOCATION=../resources/authgear/generated/manifest.json
TEMP_DIR=$(mktemp -d)

if [ $N -lt 2 ]; then
  echo "[ERROR] The build number must be greater than 1"
  exit 1
fi

build_bundles()
{
  for i in $(seq 1 $N)
  do
    echo "[INFO] Build no.$i"

    # Clear vite cache before building bundles
    rm -rf node_modules/.vite > /dev/null 2>&1
    npm run build > /dev/null 2>&1

    python3 -m json.tool --sort-keys --no-ensure-ascii --indent 2 <$MANIFEST_LOCATION >$TEMP_DIR/$i.json
  done
}

compare_bundles()
{
  echo "[INFO] Comparing $i builds..."

  for j in $(seq 2 $N)
  do
    diff $TEMP_DIR/1.json $TEMP_DIR/$j.json
    # Used to check the error code
    V=$?
    if [ $V -eq 1 ];then
      echo "[ERROR] Build hashes are not identical"
      exit $V
    fi
  done

  echo "[INFO] Everything works fine, Bye!"
}

main()
{
  build_bundles
  compare_bundles
}

echo "[INFO] Start of script"
main
echo "[INFO] End of script"
