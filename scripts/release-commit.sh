if [ -z "$SKYGEAR_VERSION" ]; then
    >&2 echo "SKYGEAR_VERSION is required."
    exit 1
fi
if [ -z "$GITHUB_TOKEN" ]; then
    >&2 echo "GITHUB_TOKEN is required."
    exit 1
fi
if [ -z "$KEY_ID" ]; then
    >&2 echo "KEY_ID is required."
    exit 1
fi
if [ -e "new-release" ]; then
    echo "Making release commit and github release..."
else
    echo "file 'new-release' is required."
    exit 1
fi

github-release release -u skygeario -r skygear-server --draft --tag v$SKYGEAR_VERSION --name "v$SKYGEAR_VERSION" --description "`cat new-release`"
cat CHANGELOG.md >> new-release && mv new-release CHANGELOG.md
sed -i "" "s/version = \".*\"/version = \"v$SKYGEAR_VERSION\"/" pkg/server/skyversion/version.go
git add CHANGELOG.md pkg/server/skyversion/version.go
git commit -m "Update CHANGELOG for v$SKYGEAR_VERSION"
git tag -a v$SKYGEAR_VERSION -s -u $KEY_ID -m "Release v$SKYGEAR_VERSION"