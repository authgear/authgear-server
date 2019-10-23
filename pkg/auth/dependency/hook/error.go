package hook

import (
	"github.com/skygeario/skygear-server/pkg/core/errors"
	skyerr "github.com/skygeario/skygear-server/pkg/core/xskyerr"
)

var WebHookDisallowed = skyerr.Forbidden.WithReason("WebHookDisallowed")

func newErrorDeliveryTimeout() error {
	return errors.New("web-hook event delivery timed out")
}

func newErrorDeliveryFailed(inner error) error {
	return errors.Newf("web-hook event delivery failed: %w", inner)
}

func newErrorDeliveryInvalidStatusCode() error {
	return errors.New("invalid status code")
}

type OperationDisallowedItem struct {
	Reason string      `json:"reason"`
	Data   interface{} `json:"data,omitempty"`
}

type disallowedErrors []OperationDisallowedItem

func (disallowedErrors) IsTagged(tag errors.DetailTag) bool { return tag == skyerr.APIErrorDetail }

func newErrorOperationDisallowed(items []OperationDisallowedItem) error {
	return WebHookDisallowed.NewWithDetails(
		"disallowed by web-hook event handler",
		map[string]interface{}{"causes": disallowedErrors(items)},
	)
}

func newErrorMutationFailed(inner error) error {
	return errors.Newf("web-hook mutation failed: %w", inner)
}
