package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	gographql "github.com/graphql-go/graphql"

	admingraphql "github.com/authgear/authgear-server/pkg/admin/graphql"
	portalgraphql "github.com/authgear/authgear-server/pkg/portal/graphql"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "must specify package: admin, portal")
		os.Exit(1)
	}

	var schema *gographql.Schema
	pkg := os.Args[1]
	switch pkg {
	case "admin":
		schema = admingraphql.Schema
	case "portal":
		schema = portalgraphql.Schema
	default:
		fmt.Fprintf(os.Stderr, "must specify package: admin, portal")
		os.Exit(1)
	}

	query, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	params := gographql.Params{
		Schema:        *schema,
		RequestString: string(query),
	}

	result := gographql.Do(params)
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		panic(err)
	}
}
