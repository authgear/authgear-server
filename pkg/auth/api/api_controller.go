package api

import (
	"context"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

//go:generate go tool mockgen -source=api_controller.go -destination=api_controller_mock_test.go -package api

// LoginLinkV1RequestSchema is the request body of POST /api/v1/login_link/inspect
// and POST /api/v1/login_link/approve.
var LoginLinkV1RequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"code": { "type": "string", "minLength": 1 }
		},
		"required": ["code"]
	}
`)

type APIControllerLoginLinkService interface {
	Inspect(ctx context.Context, code string) error
	Approve(ctx context.Context, code string) error
}

// APIController takes the request body of an API, validates it, and calls the
// service. The HTTP handlers and Auth UI both call it, so Auth UI can do only
// what the API allows.
type APIController struct {
	LoginLinks APIControllerLoginLinkService
}

// InspectLoginLink is POST /api/v1/login_link/inspect.
func (c *APIController) InspectLoginLink(ctx context.Context, body map[string]any) error {
	code, err := c.parseLoginLinkV1Code(ctx, body)
	if err != nil {
		return err
	}
	return c.LoginLinks.Inspect(ctx, code)
}

// ApproveLoginLink is POST /api/v1/login_link/approve.
func (c *APIController) ApproveLoginLink(ctx context.Context, body map[string]any) error {
	code, err := c.parseLoginLinkV1Code(ctx, body)
	if err != nil {
		return err
	}
	return c.LoginLinks.Approve(ctx, code)
}

func (c *APIController) parseLoginLinkV1Code(ctx context.Context, body map[string]any) (string, error) {
	err := LoginLinkV1RequestSchema.Validator().ValidateValueWithMessage(ctx, body, "invalid request body")
	if err != nil {
		return "", err
	}
	// The schema guarantees a non-empty string.
	return body["code"].(string), nil
}
