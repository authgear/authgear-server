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
$ export SKYGEAR_VERSION="2.1.0"

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

## Update changelog, version.go, release commit and tag
$ make release-commit

### If the release is latest (official release with the highest version number)
$ git tag -f latest && git push git@github.com:SkygearIO/skygear-server.git :latest
$ git push --follow-tags git@github.com:SkygearIO/skygear-server.git master v$SKYGEAR_VERSION latest

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
