#!/bin/sh -e
if [ -z "$SKYGEAR_VERSION" ]; then
    >&2 echo "SKYGEAR_VERSION is required."
    exit 1
fi
if [ -z "$GITHUB_TOKEN" ]; then
    >&2 echo "GITHUB_TOKEN is required."
    exit 1
fi
if [ -e "new-release" ]; then
    echo "Making release commit and github release..."
else
    >&2 echo "file 'new-release' is required."
    exit 1
fi

github-release release -u skygeario -r skygear-server --draft --tag v$SKYGEAR_VERSION --name "v$SKYGEAR_VERSION" --description "`cat new-release`"
echo "" >> new-release && cat CHANGELOG.md >> new-release && mv new-release CHANGELOG.md
make update-version SKYGEAR_VERSION=$SKYGEAR_VERSION
git add CHANGELOG.md pkg/core/apiversion/apiversion.go
git commit -m "Update CHANGELOG for v$SKYGEAR_VERSION"
git tag -a v$SKYGEAR_VERSION -s -m "Release v$SKYGEAR_VERSION"
