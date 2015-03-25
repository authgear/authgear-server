Welcome to the Oursky Deployment Project

[![Build Status](https://magnum.travis-ci.com/oursky/ourd.svg?token=TS65G314JpxpG31zryWn)](https://magnum.travis-ci.com/oursky/ourd)

Dependencies
============
1. go v1.4
2. https://github.com/tools/godep is used for managing go lib
3. PostgreSQL 9.4

Test
====
You may refer to .travis.yml

1. Create test DB `ourd_test` on local PostgreSQL
1. `go test github.com/oursky/ourd/...`

For local development, you are suggested to open GoConvey.
refs: https://github.com/smartystreets/goconvey

