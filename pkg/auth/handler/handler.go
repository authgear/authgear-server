package handler

import (
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

// nolint: deadcode
var (
	uuidNew = uuid.New
)

// nolint: deadcode
/*
	@ID EmptyResponse
	@Response
		Empty response.
		@JSONSchema
*/
const emptyResponseSchema = `
{
	"type": "object",
	"properties": {
		"result": { "type": "object" }
	}
}
`
