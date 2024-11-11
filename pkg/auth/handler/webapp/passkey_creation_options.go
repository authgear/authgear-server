package webapp

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigurePasskeyCreationOptionsRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST").
		WithPathPattern("/_internals/passkey/creation_options")
}

type PasskeyCreationOptionsService interface {
	MakeCreationOptions(ctx context.Context, userID string) (*model.WebAuthnCreationOptions, error)
}

type PasskeyCreationOptionsHandler struct {
	Page     PageService
	Database *appdb.Handle
	JSON     JSONResponseWriter
	Passkey  PasskeyCreationOptionsService
}

func (h *PasskeyCreationOptionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		if err != nil {
			h.JSON.WriteResponse(w, &api.Response{Error: err})
		}
	}()

	var creationOptions *model.WebAuthnCreationOptions
	err = h.Database.ReadOnly(r.Context(), func(ctx context.Context) error {
		webSession := webapp.GetSession(ctx)
		if webSession != nil {
			err := h.Page.PeekUncommittedChanges(ctx, webSession, func(graph *interaction.Graph) error {
				userID := graph.MustGetUserID()
				var err error
				creationOptions, err = h.Passkey.MakeCreationOptions(ctx, userID)
				if err != nil {
					return err
				}

				return nil
			})
			if err != nil {
				return err
			}
		} else {
			userID := session.GetUserID(ctx)
			if userID == nil {
				return apierrors.NewBadRequest("session not found")
			}

			var err error
			creationOptions, err = h.Passkey.MakeCreationOptions(ctx, *userID)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return
	}

	h.JSON.WriteResponse(w, &api.Response{
		Result: creationOptions,
	})
}
