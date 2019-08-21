#!/bin/sh -eu

VERSION="$(git describe --always)"
PROJ_ROOT="$PWD"
TMP_DIR="$(mktemp -d)"

git branch -C origin/gh-pages gh-pages

cd "$TMP_DIR"

go get -u github.com/skygeario/openapi3-gen/cmd/openapi3-gen

git clone --no-checkout "$PROJ_ROOT" gh-pages
cd gh-pages
git checkout -t origin/gh-pages || git checkout --orphan gh-pages
git reset --hard

mkdir -p apis/auth
make -C "$PROJ_ROOT" -e DOC_PATH="$PWD/apis/auth/$VERSION.yaml" generate-doc-auth

URL="https://generator.swagger.io/?url=https://skygeario.github.io/skygear-server/apis/auth/$VERSION.yaml"
sed "s|{URL}|$URL|g" "$PROJ_ROOT/scripts/template/index.html" > index.html

git add .
if git commit -m "Update documentation for $VERSION"; then
    git push origin gh-pages
    echo "gh-pages branch updated in $PROJ_ROOT"
else
    echo "gh-pages branch is up to update"
fi

rm -rf "$TMP_DIR"
