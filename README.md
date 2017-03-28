![Skygear Logo](.github/skygear-logo.png)

Skygear Server is a cloud backend for making web, mobile and IoT app development easier.

It provides the following features for common app development:
* Skygear Auth - User Authentication, Social Login, Security
* Skygear Cloud DB - Document-like API/SDK with auto-migration based on PostgreSQL
* Skygear Chat - Build full-featured messaging into your app fast
  [Server](https://github.com/SkygearIO/chat), [JS
  SDK](https://github.com/SkygearIO/chat-SDK-JS), [Android
  SDK](https://github.com/SkygearIO/chat-SDK-Android), [iOS
  SDK](https://github.com/SkygearIO/chat-SDK-iOS)
* Skygear Cloud Functions - Run event-trigger functions without thinking about
  server
* Skygear CMS - Drop-in CMS for business users integrated with Skygear Auth and
  Cloud DB [Opensource soon](https://github.com/oursky/skygear-cms)
* Skygear Pubsub - Cloud Pub/sub API for real-time and reliable messaging and
  connection.
* Skygear Push - Push Notification

[Skygear.io](https://skygear.io) is the commercial hosted platform for
developers who don't want to manage their own infrastructure.

[![Slack](https://img.shields.io/badge/join-Slack-green.svg)](https://skygear.signup.team/)
[![Google Group](https://img.shields.io/badge/join-Google%20Group-green.svg)](https://groups.google.com/forum/#!forum/skygear-user-group)

[![Build Status](https://travis-ci.org/SkygearIO/skygear-server.svg?branch=master)](https://travis-ci.org/SkygearIO/skygear-server)
[![Go Report Card](https://goreportcard.com/badge/github.com/skygeario/skygear-server)](https://goreportcard.com/report/github.com/skygeario/skygear-server)

## Getting Started

To get started, you need to have a Skygear Server and include one of the SDKs into your app.

The easiest way to start using Skygear is sign-up a free Development Plan at [https://portal.skygear.io](https://portal.skygear.io)

Read the [Skygear Server Guide](https://docs.skygear.io/guides/advanced/server/)
for local deployment. Download binaries of [Skygear Releases here](https://github.com/SkygearIO/skygear-server/releases)


## Connect your app to Skygear Server
Skygear provides SDKs for all the major platforms. Please refer to the guide for each platform to learn how to connect your app to Skygear Server: [iOS](https://docs.skygear.io/guides/get-started/ios/) / [Android](https://docs.skygear.io/guides/get-started/android/) / [JavaScript](https://docs.skygear.io/guides/get-started/js/)

## Documentation
* [Skygear Guides](https://docs.skygear.io/guides/)
* [Skygear API References](https://docs.skygear.io/api-reference/)
* [Sample Projects and Tutorials](https://github.com/skygear-demo)

### Can I Access The Docs Offline?

The [skygear-doc repository](https://github.com/skygeario/skygear-doc) is public and all the content files are in markdown. If you'd like to keep a copy locally, please do!

## Support

If you believe you've found an bugs or feature requests for Skygear, please feel
free to [report an issue](https://github.com/SkygearIO/skygear-server/issues).

Please do not use the issue tracker for personal support requests. Instead, use
[Stack Overflow](http://stackoverflow.com/questions/tagged/skygear) community,
join our [Google Group](https://groups.google.com/forum/#!forum/skygear-user-group),
or [Slack Team](https://skygear.signup.team) to chat with our
supports/developers.

For issues about [Skygear Guides](https://docs.skygear.io) please use the
[Skygear Guides issues tracker](https://github.com/skygeario/guides/issues).

## Configuration

Skygear is configure via environment variable. It also support `.env` file for
easy development.

The minimal configuration will be provide `API_KEY` and `MASTER_KEY`

```shell
$ API_KEY=changeme MASTER_KEY=secret ./skygear-server
```

Check out [`.env`](https://github.com/SkygearIO/skygear-server/blob/master/.env.example)
for configuration reference. Once you configure the `.env`
correctly, you can simple kick start the server by following.

```shell
$ ./skygear-server
```

## How to contribute

Pull Requests Welcome!

We really want to see Skygear grows and thrives in the open source community.
If you have any fixes or suggestions, simply send us a pull request!

### Dependencies

* Golang 1.7
* PostgreSQL 9.5 with PostGIS extension
* Redis
* libsodium, zeromq and czmq if using ZeroMQ as a plugin transport

If using Mac OS X, you can get the ZeroMQ dependencies using Homebrew:

```shell
$ brew install libsodium zeromq czmq
```

### Building from source

```shell
$ go get github.com/Masterminds/glide
$ make vendor
$ # export WITH_ZMQ=1 # If you need ZeroMQ support
$ make build
```

#### Building with Nix

Assuming you have [Nix](https://nixos.org/nix/) installed,
Skygear can be built with the following command:

```shell
nix-build default.nix
```

You will have a symbolic link `result-bin` linking to the binary.

### Testing

1. Create a PostgreSQL database called `skygear_test` with PostGIS enabled:

```shell
psql -h db -c 'CREATE DATABASE skygear_test;' -U postgres
psql -h db -c 'CREATE EXTENSION postgis;' -U postgres -d skygear_test
psql -h db -c 'CREATE EXTENSION citext;' -U postgres -d skygear_test
```

2. Test case assume the 127.0.0.1 have access to `skygear_test`, add the
following to `pg_hba.conf`:

```
host    all             all             127.0.0.1/32            trust
```

3. Install golang packages required for testing (check `.travis.yml` for the
   list).

4. Run `go test github.com/skygeario/skygear-server/pkg/...`.

You can also run the test suite in Docker:

```
$ make vendor WITH_DOCKER=1  # install dependencies
$ make before-docker-test    # start dependent services
$ make test WITH_DOCKER=1    # run test
$ make after-docker-test     # clean up docker containers
```

## License & Copyright

```
Copyright (c) 2015-present, Oursky Ltd.
All rights reserved.

This source code is licensed under the Apache License version 2.0
found in the LICENSE file in the root directory of this source tree.
An additional grant of patent rights can be found in the PATENTS
file in the same directory.

```
