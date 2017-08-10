- Upload `py-skygear` to PyPI as `skygear`
- Docker Hub automatically build `skygeario/py-skygear` triggered by git push
- Upload `skygear-SDK-iOS` to [CocoaPods](https://cocoapods.org/pods/SKYKit) as `SKYKit`
- Upload `skygear-SDK-JS` to [npm](https://www.npmjs.com/package/skygear) as `skygear`
- Upload `skygear-SDK-Android` to [jcenter](https://bintray.com/skygeario/maven/skygear-android)

## Version number

Version number conforms to `x.y.z` format. For each release cycle, `y`
is incremented by 1. The same `x.y` corresponds to the tag of skygear and it
will stay the same for all above modules during the release cycle.
For hot fix within a release cycle, increment `z`
by 1 and `z` is reset to 0 on the next release cycle.

## How to release?

### Preparation

```shell
$ export GITHUB_TOKEN=abcdef1234 # Need Repos scope for update release notes
$ export SKYGEAR_VERSION="0.5.0"
$ export KEY_ID="12CDA17C"

$ brew install github-release
$ brew install gpg2
```

*IMPORTANT*: This guide assumes that your `origin` points to
`skygeario/skygear-server`. Make sure you are on `master` branch and the
branch is the same as the `origin/master`.

### skygear-server

```shell
## Draft new release changelog
$ git log --first-parent `git describe --abbrev=0`.. > new-release
$ edit new-release
$ github-release release -u skygeario -r skygear-server --draft --tag v$SKYGEAR_VERSION --name "v$SKYGEAR_VERSION" --description "`cat new-release`"

## Update changelog
$ cat CHANGELOG.md >> new-release && mv new-release CHANGELOG.md
$ sed -i "" "s/version = \".*\"/version = \"v$SKYGEAR_VERSION\"/" pkg/server/skyversion/version.go
$ git add CHANGELOG.md pkg/server/skyversion/version.go
$ git commit -m "Update CHANGELOG for v$SKYGEAR_VERSION"

## Tag commit
$ git tag -a v$SKYGEAR_VERSION -s -u $KEY_ID -m "Release v$SKYGEAR_VERSION"

### If the release is latest (official release with the highest version number)
$ git tag -f latest
$ git push origin :latest

$ git push --follow-tags origin master v$SKYGEAR_VERSION latest

## Click `Publish release` in github release page
```

### py-skygear

```shell
## Draft new release changelog
$ git log --first-parent `git describe --abbrev=0`.. > new-release
$ edit new-release
$ github-release release -u skygeario -r py-skygear --draft --tag v$SKYGEAR_VERSION --name "v$SKYGEAR_VERSION" --description "`cat new-release`"

## Update changelog and version number
$ cat CHANGELOG.md >> new-release && mv new-release CHANGELOG.md
$ make update-version VERSION=$SKYGEAR_VERSION
$ git add CHANGELOG.md setup.py skygear/__version__.py
$ git commit -m "Update CHANGELOG for v$SKYGEAR_VERSION"

## Release to pypi (Only for official release)
$ python3 setup.py sdist upload

## Tag and push commit
$ git tag -a v$SKYGEAR_VERSION -s -u $KEY_ID -m "Release v$SKYGEAR_VERSION"

### If the release is latest (official release with the highest version number)
$ git tag -f latest
$ git push origin :latest

$ git push --follow-tags origin master v$SKYGEAR_VERSION latest


## Click `Publish release` in github release page
```

### skygear-SDK-iOS

**IMPORTANT**: Note that CocoaPods does not allow tag prefixed with `v`.
Therefore the tag name is different from other projects.

**IMPORTANT**: CocoaPods requires that that tag is pushed to repository before
it will accept a new release.

```shell
## Draft new release changelog
$ git log --first-parent `git describe --abbrev=0`.. > new-release
$ edit new-release
$ github-release release -u skygeario -r skygear-SDK-iOS --draft --tag $SKYGEAR_VERSION --name "$SKYGEAR_VERSION" --description "`cat new-release`"

## Update changelog and version number
$ cat CHANGELOG.md >> new-release && mv new-release CHANGELOG.md
$ make update-version VERSION=$SKYGEAR_VERSION
$ git add CHANGELOG.md SKYKit.podspec Pod/Classes/SKYKit+version.h
$ git commit -m "Update CHANGELOG for $SKYGEAR_VERSION"

## Tag and push commit
$ git tag -a $SKYGEAR_VERSION -s -u $KEY_ID -m "Release $SKYGEAR_VERSION"

### If the release is latest (official release with the highest version number)
$ git tag -f latest
$ git push origin :latest

$ git push --follow-tags origin master $SKYGEAR_VERSION latest

## Push commit to Cocoapods (Only for official release)
$ pod trunk push SKYKit.podspec --allow-warnings

## Click `Publish release` in github release page
```

### skygear-SDK-JS

```shell
## Draft new release changelog
$ git log --first-parent `git describe --abbrev=0`.. > new-release
$ edit new-release
$ github-release release -u skygeario -r skygear-SDK-JS --draft --tag v$SKYGEAR_VERSION --name "v$SKYGEAR_VERSION" --description "`cat new-release`"

## Update changelog
$ cat CHANGELOG.md >> new-release && mv new-release CHANGELOG.md
$ git add CHANGELOG.md

## Changing the version number and releasing all packages to npm using lerna.
$ make update-version VERSION=$SKYGEAR_VERSION
$ npm run lerna bootstrap # make sure dependencies are linked
$ npm run prepublish # Build all packages
$ npm run lerna publish -- --skip-git --repo-version $SKYGEAR_VERSION # Publish all packages
## For publishing to alpha channel
## $ npm run lerna publish -- --skip-git --npm-tag=alpha --repo-version $SKYGEAR_VERSION

## Tag and push commit
$ git add CHANGELOG.md lerna.json gulp/context.js packages/*/package.json
$ git commit -m "Update CHANGELOG for $SKYGEAR_VERSION"
$ git tag -a v$SKYGEAR_VERSION -s -u $KEY_ID -m "Release v$SKYGEAR_VERSION"

### If the release is latest (official release with the highest version number)
$ git tag -f latest
$ git push origin :latest

$ git push --follow-tags origin master v$SKYGEAR_VERSION latest

## Click `Publish release` in github release page
```

### skygear-SDK-Android

```shell
## Draft new release notes
$ git log --first-parent `git describe --abbrev=0`.. > new-release
$ edit new-release
$ github-release release -u skygeario -r skygear-SDK-Android --draft --tag $SKYGEAR_VERSION --name "$SKYGEAR_VERSION" --description "`cat new-release`"

## Update changelog
$ cat CHANGELOG.md >> new-release && mv new-release CHANGELOG.md
$ make update-version VERSION=$SKYGEAR_VERSION
$ git add CHANGELOG.md skygear/build.gradle
$ git commit -m "Update CHANGELOG for $SKYGEAR_VERSION"

## Tag and push commit
$ git tag -a $SKYGEAR_VERSION -s -u $KEY_ID -m "Release $SKYGEAR_VERSION"

### If the release is latest (official release with the highest version number)
$ git tag -f latest
$ git push origin :latest

$ git push --follow-tags origin master $SKYGEAR_VERSION latest

## Click `Publish release` in github release page
```

### skycli

```shell
## Draft new release changelog
$ git log --first-parent `git describe --abbrev=0`.. > new-release
$ edit new-release
$ github-release release -u skygeario -r skycli --draft --tag v$SKYGEAR_VERSION --name "v$SKYGEAR_VERSION" --pre-release --description "`cat new-release`"

## Update changelog
$ cat CHANGELOG.md >> new-release && mv new-release CHANGELOG.md
$ git add CHANGELOG.md
$ git commit -m "Update CHANGELOG for v$SKYGEAR_VERSION"

## Tag and push commit
$ git tag -a v$SKYGEAR_VERSION -s -u $KEY_ID -m "Release v$SKYGEAR_VERSION"
$ git push --follow-tags origin v$SKYGEAR_VERSION
$ git push origin

## Click `Publish release` in github release page
```

## Other notes

To show tag info:

```shell
$ git show v0.5.0
```


To delete tags:

```
$ git tag -d v$SKYGEAR_VERSION
$ git push origin :refs/tags/v$SKYGEAR_VERSION
```
