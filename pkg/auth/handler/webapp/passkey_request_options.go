package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigurePasskeyRequestOptionsRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST").
		WithPathPattern("/_internals/passkey/request_options")
}

type PasskeyRequestOptionsService interface {
	MakeConditionalRequestOptions() (*model.WebAuthnRequestOptions, error)
	MakeModalRequestOptions() (*model.WebAuthnRequestOptions, error)
	MakeModalRequestOptionsWithUser(userID string) (*model.WebAuthnRequestOptions, error)
}

type PasskeyRequestOptionsHandler struct {
	Page     PageService
	Database *appdb.Handle
	JSON     JSONResponseWriter
	Passkey  PasskeyRequestOptionsService
}

func (h *PasskeyRequestOptionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		if err != nil {
			h.JSON.WriteResponse(w, &api.Response{Error: err})
		}
	}()

	err = r.ParseForm()
	if err != nil {
		return
	}

	conditional := r.FormValue("conditional") == "true"
	allowCredentials := r.FormValue("allow_credentials") == "true"

	var requestOptions *model.WebAuthnRequestOptions
	err = h.Database.ReadOnly(func() error {
		if conditional {
			requestOptions, err = h.Passkey.MakeConditionalRequestOptions()
			if err != nil {
				return err
			}
			return nil
		}

		if allowCredentials {
			session := webapp.GetSession(r.Context())
			if session == nil {
				err = apierrors.NewBadRequest("session not found")
				return err
			}
			err := h.Page.PeekUncommittedChanges(session, func(graph *interaction.Graph) error {
				userID := graph.MustGetUserID()
				var err error
				requestOptions, err = h.Passkey.MakeModalRequestOptionsWithUser(userID)
				if err != nil {
					return err
				}

				return nil
			})
			if err != nil {
				return err
			}
			return nil
		}

		requestOptions, err = h.Passkey.MakeModalRequestOptions()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return
	}

	h.JSON.WriteResponse(w, &api.Response{
		Result: requestOptions,
	})
}
