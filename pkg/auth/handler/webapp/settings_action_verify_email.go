package webapp

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureSettingsActionVerifyEmailRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/_internals/settings_action/verify_email")
}

type SettingsActionVerifyEmailHandler struct {
	ControllerFactory ControllerFactory
	Identities        SettingsIdentityService
	Verification      SettingsVerificationService
	ErrorCookie       ErrorCookie
}

func (h *SettingsActionVerifyEmailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	getIdentityToVerify := func(userID string) (*identity.Info, error) {
		identities, err := h.Identities.ListByUser(userID)
		if err != nil {
			return nil, err
		}
		var emailLoginIDIdentities []*identity.Info
		for _, i := range identities {
			if i.Type == model.IdentityTypeLoginID &&
				i.LoginID.LoginIDType == model.LoginIDKeyTypeEmail {
				ii := i
				emailLoginIDIdentities = append(emailLoginIDIdentities, ii)
			}
		}

		verificationStatuses, err := h.Verification.GetVerificationStatuses(emailLoginIDIdentities)
		if err != nil {
			return nil, err
		}

		for _, i := range emailLoginIDIdentities {
			cvs := verificationStatuses[i.ID]
			if len(cvs) == 0 {
				continue
			}
			claimVerificationStatus := cvs[0]
			if !claimVerificationStatus.Verified && claimVerificationStatus.EndUserTriggerable {
				return i, nil
			}
		}

		return nil, errors.New("no identity to verify")
	}

	_, hasErr := h.ErrorCookie.GetError(r)
	if hasErr {
		http.Redirect(w, r, "/errors/error", http.StatusFound)
		return
	}

	ctrl.Get(func() error {
		userID := session.GetUserID(r.Context())
		if userID == nil {
			// fixme: handle error
			return errors.New("login required")
		}
		identity, err := getIdentityToVerify(*userID)
		if err != nil {
			return err
		}
		opts := webapp.SessionOptions{}
		intent := intents.NewIntentVerifyIdentity(*userID, model.IdentityTypeLoginID, identity.ID)
		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
			input = nil
			return
		})
		if err != nil {
			return err
		}
		result.WriteResponse(w, r)
		return nil
	})
}
