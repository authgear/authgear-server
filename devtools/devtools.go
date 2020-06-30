package devtools

// The imports here are main packages.
// So they are not importable.
// We list them here so that `go mod`
// know them and add them to go.mod
// Because the imports are main packages,
// this package itself cannot compile.
import (
	_ "github.com/golang/mock/mockgen"
	_ "github.com/google/wire/cmd/wire"
)
