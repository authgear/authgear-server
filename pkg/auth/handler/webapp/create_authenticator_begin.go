package webapp

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureCreateAuthenticatorBeginRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/create_authenticator_begin")
}

type CreateAuthenticatorBeginSetupTOTP struct{}

var _ nodes.InputCreateAuthenticatorTOTPSetup = &CreateAuthenticatorBeginSetupTOTP{}

func (i *CreateAuthenticatorBeginSetupTOTP) SetupTOTP() {}

type CreateAuthenticatorBeginNode interface {
	GetCreateAuthenticatorEdges() ([]newinteraction.Edge, error)
}

type CreateAuthenticatorBeginHandler struct {
	Database *db.Handle
	WebApp   WebAppService
}

func (h *CreateAuthenticatorBeginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var err error

	edgeIndexString := r.Form.Get("x_edge")
	if edgeIndexString == "" {
		edgeIndexString = "0"
	}
	edgeIndex, err := strconv.Atoi(edgeIndexString)
	if err != nil {
		edgeIndex = 0
	}

	var state *webapp.State
	var graph *newinteraction.Graph

	h.Database.WithTx(func() error {
		state, graph, err = h.WebApp.Get(StateID(r))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		return nil
	})

	var node CreateAuthenticatorBeginNode
	if !graph.FindLastNode(&node) {
		panic("create_authenticator_begin: expected graph has node implementing CreateAuthenticatorBeginNode")
	}
	edges, err := node.GetCreateAuthenticatorEdges()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if edgeIndex >= len(edges) {
		edgeIndex = 0
	}

	h.Database.WithTx(func() error {
		selectedEdge := edges[edgeIndex]
		switch selectedEdge := selectedEdge.(type) {
		case *nodes.EdgeCreateAuthenticatorPassword:
			http.Redirect(w, r, webapp.AttachStateID(state.ID, &url.URL{
				Path: "/create_password",
			}).String(), http.StatusFound)
		case *nodes.EdgeCreateAuthenticatorOOBSetup:
			http.Redirect(w, r, webapp.AttachStateID(state.ID, &url.URL{
				Path: "/setup_oob_otp",
			}).String(), http.StatusFound)
		case *nodes.EdgeCreateAuthenticatorTOTPSetup:
			result, err := h.WebApp.PostInput(StateID(r), func() (input interface{}, err error) {
				input = &CreateAuthenticatorBeginSetupTOTP{}
				return
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}
			result.WriteResponse(w, r)
		default:
			panic(fmt.Errorf("webapp: unexpected edge: %T", selectedEdge))
		}

		return nil
	})
}

type CreateAuthenticatorAlternative struct {
	Type string
	URL  string
}

func DeriveCreateAuthenticatorAlternatives(stateID string, graph *newinteraction.Graph, currentType authn.AuthenticatorType) (alternatives []CreateAuthenticatorAlternative, err error) {
	var node CreateAuthenticatorBeginNode
	if !graph.FindLastNode(&node) {
		panic("create_authenticator_begin: expected graph has node implementing CreateAuthenticatorBeginNode")
	}

	edges, err := node.GetCreateAuthenticatorEdges()
	if err != nil {
		return
	}

	for i, edge := range edges {
		q := url.Values{}
		q.Set("x_edge", strconv.Itoa(i))

		var typ authn.AuthenticatorType
		switch edge.(type) {
		case *nodes.EdgeCreateAuthenticatorPassword:
			typ = authn.AuthenticatorTypePassword
		case *nodes.EdgeCreateAuthenticatorOOBSetup:
			typ = authn.AuthenticatorTypeOOB
		case *nodes.EdgeCreateAuthenticatorTOTPSetup:
			typ = authn.AuthenticatorTypeTOTP
		default:
			panic(fmt.Errorf("create_authenticator_begin: unexpected edge: %T", edge))
		}

		if typ != currentType {
			alternatives = append(alternatives, CreateAuthenticatorAlternative{
				Type: string(typ),
				URL: webapp.AttachStateID(stateID, &url.URL{
					Path:     "/create_authenticator_begin",
					RawQuery: q.Encode(),
				}).String(),
			})
		}
	}

	return
}
