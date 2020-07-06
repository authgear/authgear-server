package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/authgear/authgear-server/pkg/auth/config"
)

type jsonSchema interface {
	DumpSchemaString(pretty bool) (string, error)
}

func main() {
	output := flag.String("o", "", "output path")
	schemaType := flag.String("s", "", "schema type")
	flag.Parse()

	var schema jsonSchema
	switch *schemaType {
	case "app-config":
		schema = config.Schema
	case "secrets-config":
		schema = config.SecretConfigSchema
	default:
		panic("unknown schema type: " + *schemaType)
	}

	json, err := schema.DumpSchemaString(true)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(*output, []byte(json), 0666)
	if err != nil {
		panic(err)
	}

	log.Printf("schema %s written to %s", *schemaType, *output)
}
