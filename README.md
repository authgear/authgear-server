![Skygear Logo](.github/skygear-logo.png)

Skygear Server is a cloud backend for making web and mobile app development easier. [https://skygear.io](https://skygear.io)



[![Build Status](https://travis-ci.org/SkygearIO/skygear-server.svg?branch=master)](https://travis-ci.org/SkygearIO/skygear-server)

## Getting Started

To get started, you need to install Skygear Server and include one of the SDKs into your app. You can see detailed procedure at the getting started guide at [https://docs.skygear.io/server/guide](https://docs.skygear.io/server/guide).

The fastest way to get Skygear Server running is to download the runnable binaries of the latest release at [https://github.com/SkygearIO/skygear-server/releases](https://github.com/SkygearIO/skygear-server/releases)

You can also sign up the Skygear Hosting at the Skygear Developer Portal at [https://portal.skygear.io](https://portal.skygear.io)

## Connect your app to Skygear Server
Skygear provides SDKs for all the major platforms. Please refer to the guide for each platform to learn how to connect your app to Skygear Server: [iOS] (https://docs.skygear.io/ios/guide) / [Android](https://docs.skygear.io/android/guide) / [JavaScript](https://docs.skygear.io/js/guide)

## Documentation
The full documentation for Skygear Server is available on our docs site. The [Skygear Server guide](https://docs.skygear.io/server/guide) is a good place to get started.

### Can I Access The Docs Offline?

The [documentation repository](https://github.com/skygeario/skygear-doc) is public and all the content files are in markdown. If you'd like to keep a copy locally, please do!

## Support

For implementation related questions or technical support, please refer to the [Stack Overflow](http://stackoverflow.com/questions/tagged/skygear) community.

If you believe you've found an issue with Skygear Server, please feel free to [report an issue](https://github.com/SkygearIO/skygear-server/issues).

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

* Golang 1.6
* PostgreSQL 9.4 with PostGIS extension
* Redis
* libsodium, zeromq and czmq if using ZeroMQ as a plugin transport

If using Mac OS X, you can get the ZeroMQ dependencies using Homebrew:

```shell
$ brew install libsodium zeromq czmq
```

### Building from source

```shell
$ go get github.com/tools/godep
$ godep restore
$ go build  # or `go build --tags zmq` for ZeroMQ support
```

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

4. Run `go test github.com/skygeario/skygear-server/...`.

## License & Copyright

```
Copyright (c) 2015-present, Oursky Ltd.
All rights reserved.

This source code is licensed under the Apache License version 2.0 
found in the LICENSE file in the root directory of this source tree. 
An additional grant of patent rights can be found in the PATENTS 
file in the same directory.

```
