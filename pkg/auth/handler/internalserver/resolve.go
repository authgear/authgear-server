package internalserver

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/log"
)

func ConfigureResolveRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("HEAD", "GET").
		WithPathPattern("/resolve")
}

//go:generate mockgen -source=resolve.go -destination=resolve_mock_test.go -package internalserver

type IdentityService interface {
	ListByUser(userID string) ([]*identity.Info, error)
}

type VerificationService interface {
	IsUserVerified(identities []*identity.Info, userID string) (bool, error)
}

type ResolveHandlerLogger struct{ *log.Logger }

func NewResolveHandlerLogger(lf *log.Factory) ResolveHandlerLogger {
	return ResolveHandlerLogger{lf.New("resolve-handler")}
}

type ResolveHandler struct {
	Identities   IdentityService
	Verification VerificationService
	Logger       ResolveHandlerLogger
}

func (h *ResolveHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	info, err := h.resolve(r)
	if err != nil {
		h.Logger.WithError(err).Error("failed to resolve user")

		http.Error(rw, "internal error", http.StatusInternalServerError)
		return
	}
	info.PopulateHeaders(rw)
}

func (h *ResolveHandler) resolve(r *http.Request) (*authn.Info, error) {
	valid := auth.IsValidAuthn(r.Context())
	userID := auth.GetUserID(r.Context())
	session := auth.GetSession(r.Context())

	var info *authn.Info
	if valid && userID != nil && session != nil {
		identities, err := h.Identities.ListByUser(*userID)
		if err != nil {
			return nil, err
		}

		isAnonymous := false
		for _, i := range identities {
			if i.Type == authn.IdentityTypeAnonymous {
				isAnonymous = true
				break
			}
		}

		isVerified, err := h.Verification.IsUserVerified(identities, *userID)
		if err != nil {
			return nil, err
		}

		info = authn.NewAuthnInfo(session.AuthnAttrs(), isAnonymous, isVerified)
	} else if !valid {
		info = &authn.Info{IsValid: false}
	}

	return info, nil
}
