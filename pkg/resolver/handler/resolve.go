package handler

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func ConfigureResolveRoute(route httproute.Route) []httproute.Route {
	route = route.WithMethods("HEAD", "GET")
	return []httproute.Route{
		route.WithPathPattern("/resolve"),
		route.WithPathPattern("/_resolver/resolve"),
	}
}

//go:generate mockgen -source=resolve.go -destination=resolve_mock_test.go -package handler

type IdentityService interface {
	ListByUser(userID string) ([]*identity.Info, error)
}

type VerificationService interface {
	IsUserVerified(identities []*identity.Info) (bool, error)
}

type Database interface {
	ReadOnly(func() error) error
}

type ResolveHandlerLogger struct{ *log.Logger }

func NewResolveHandlerLogger(lf *log.Factory) ResolveHandlerLogger {
	return ResolveHandlerLogger{lf.New("resolve-handler")}
}

type ResolveHandler struct {
	Database     Database
	Identities   IdentityService
	Verification VerificationService
	Logger       ResolveHandlerLogger
}

func (h *ResolveHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	_ = h.Database.ReadOnly(func() error {
		return h.Handle(rw, r)
	})
}

func (h *ResolveHandler) Handle(rw http.ResponseWriter, r *http.Request) (err error) {
	info, err := h.resolve(r)
	if err != nil {
		h.Logger.WithError(err).Error("failed to resolve user")

		http.Error(rw, "internal error", http.StatusInternalServerError)
		return
	}
	if info != nil {
		info.PopulateHeaders(rw)
	}

	return
}

func (h *ResolveHandler) resolve(r *http.Request) (*model.SessionInfo, error) {
	valid := session.HasValidSession(r.Context())
	userID := session.GetUserID(r.Context())
	s := session.GetSession(r.Context())

	var info *model.SessionInfo
	if valid && userID != nil && s != nil {
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

		isVerified, err := h.Verification.IsUserVerified(identities)
		if err != nil {
			return nil, err
		}

		info = session.NewInfo(s, isAnonymous, isVerified)
	} else if !valid {
		info = &model.SessionInfo{IsValid: false}
	}

	return info, nil
}
