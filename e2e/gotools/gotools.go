//go:build tools

package gotools

// As of go1.21, we need to build flag to enable this trick.
// See https://github.com/golang/go/issues/48429
// and https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
//
// The imports here are main packages.
// So they are not importable.
// We list them here so that `go mod`
// know them and add them to go.mod
// Because the imports are main packages,
// this package itself cannot compile.
import (
	_ "github.com/google/wire/cmd/wire"
)
