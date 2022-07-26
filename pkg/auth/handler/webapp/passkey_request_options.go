package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigurePasskeyRequestOptionsRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST").
		WithPathPattern("/passkey/request_options")
}

type PasskeyRequestOptionsService interface {
	MakeConditionalRequestOptions() (*model.WebAuthnRequestOptions, error)
	MakeModalRequestOptions(userID string) (*model.WebAuthnRequestOptions, error)
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

	var requestOptions *model.WebAuthnRequestOptions
	err = h.Database.ReadOnly(func() error {
		if conditional {
			requestOptions, err = h.Passkey.MakeConditionalRequestOptions()
			if err != nil {
				return err
			}
		} else {
			session := webapp.GetSession(r.Context())
			if session == nil {
				return webapp.ErrSessionNotFound
			}
			err := h.Page.PeekUncommittedChanges(session, func(graph *interaction.Graph) error {
				userID := graph.MustGetUserID()
				var err error
				requestOptions, err = h.Passkey.MakeModalRequestOptions(userID)
				if err != nil {
					return err
				}

				return nil
			})
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
		Result: requestOptions,
	})
}
