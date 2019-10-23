package handler

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

// nolint: deadcode
var (
	uuidNew = uuid.New
	timeNow = func() time.Time { return time.Now().UTC() }
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
