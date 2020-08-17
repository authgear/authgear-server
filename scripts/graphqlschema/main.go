package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	gographql "github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/admin/graphql"
)

func main() {
	query, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	params := gographql.Params{
		Schema:        *graphql.Schema,
		RequestString: string(query),
	}

	result := gographql.Do(params)
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		panic(err)
	}
}
