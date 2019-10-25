#!/bin/sh -eu

VERSION="$(git describe --always)"
PROJ_ROOT="$PWD"
TMP_DIR="$(mktemp -d)"

git fetch origin gh-pages:gh-pages -f

cd "$TMP_DIR"

go get -u github.com/skygeario/openapi3-gen/cmd/openapi3-gen

git clone --no-checkout "$PROJ_ROOT" gh-pages
cd gh-pages
git checkout -t origin/gh-pages || git checkout --orphan gh-pages
git reset --hard

mkdir -p apis/auth
mkdir -p apis/asset

make -C "$PROJ_ROOT" -e DOC_PATH="$PWD/apis/auth/$VERSION.yaml" generate-doc-auth
make -C "$PROJ_ROOT" -e DOC_PATH="$PWD/apis/asset/$VERSION.yaml" generate-doc-asset

AUTH_URL="https://generator.swagger.io/?url=https://skygeario.github.io/skygear-server/apis/auth/$VERSION.yaml"
ASSET_URL="https://generator.swagger.io/?url=https://skygeario.github.io/skygear-server/apis/asset/$VERSION.yaml"

sed "s|{AUTH_URL}|$AUTH_URL|g" "$PROJ_ROOT/scripts/template/index.html" > index.html
sed "s|{ASSET_URL}|$ASSET_URL|g" "$PROJ_ROOT/scripts/template/index.html" > index.html

git add .
if git commit -m "Update documentation for $VERSION"; then
    git push origin gh-pages
    echo "gh-pages branch updated in $PROJ_ROOT"
else
    echo "gh-pages branch is up to update"
fi

rm -rf "$TMP_DIR"
