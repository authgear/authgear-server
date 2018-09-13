package authz

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/model"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func DenyNoAccessKey(r *http.Request, ctx handler.AuthContext) error {
	keyType := ctx.AccessKeyType
	if keyType == model.NoAccessKey {
		return skyerr.NewError(skyerr.AccessKeyNotAccepted, "api key required")
	}

	return nil
}

func RequireMasterKey(r *http.Request, ctx handler.AuthContext) error {
	keyType := ctx.AccessKeyType
	if keyType != model.MasterAccessKey {
		return skyerr.NewError(skyerr.AccessKeyNotAccepted, "master key required")
	}

	return nil
}
