package loader

import (
	"context"
)

type AuthzService interface {
	CheckAccessOfViewer(ctx context.Context, appID string) (userID string, err error)
}
