package main

import (
	"os"

	testrunner "github.com/authgear/authgear-server/e2e/pkg/testrunner"
)

func main() {
	schema, err := testrunner.DumpSchema()
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("./schema.json", []byte(schema), 0644)
	if err != nil {
		panic(err)
	}
}
