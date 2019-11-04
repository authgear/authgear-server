package validation

import (
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

func JSONPointer(tokens ...interface{}) string {
	var ctx *gojsonschema.JsonContext
	for _, token := range tokens {
		ctx = gojsonschema.NewJsonContext(fmt.Sprint(token), ctx)
	}
	return ctx.JSONPointer()
}
