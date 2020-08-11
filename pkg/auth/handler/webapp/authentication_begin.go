package webapp

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
)

func ConfigureAuthenticationBeginRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/authentication_begin")
}

type AuthenticationBeginInput struct {
	AuthenticatorIndex int
}

var _ nodes.InputAuthenticationOOBTrigger = &AuthenticationBeginInput{}

func (i *AuthenticationBeginInput) GetOOBAuthenticatorIndex() int {
	return i.AuthenticatorIndex
}

type AuthenticationBeginNode interface {
	GetAuthenticationEdges() []newinteraction.Edge
}

type AuthenticationBeginHandler struct {
	Database *db.Handle
	WebApp   WebAppService
}

func (h *AuthenticationBeginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	authenticatorIndexString := r.Form.Get("x_authenticator")
	if authenticatorIndexString == "" {
		authenticatorIndexString = "0"
	}
	authenticatorIndex, err := strconv.Atoi(authenticatorIndexString)
	if err != nil {
		authenticatorIndex = 0
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

	var node AuthenticationBeginNode
	if !graph.FindLastNode(&node) {
		panic("authentication_begin: expected graph has node implementing AuthenticationBeginNode")
	}
	edges := node.GetAuthenticationEdges()

	if edgeIndex >= len(edges) {
		edgeIndex = 0
	}

	h.Database.WithTx(func() error {
		selectedEdge := edges[edgeIndex]
		switch selectedEdge := selectedEdge.(type) {
		case *nodes.EdgeAuthenticationPassword:
			http.Redirect(w, r, webapp.AttachStateID(state.ID, &url.URL{
				Path: "/enter_password",
			}).String(), http.StatusFound)
		case *nodes.EdgeAuthenticationTOTP:
			http.Redirect(w, r, webapp.AttachStateID(state.ID, &url.URL{
				Path: "/enter_totp",
			}).String(), http.StatusFound)
		case *nodes.EdgeAuthenticationOOBTrigger:
			if authenticatorIndex >= len(selectedEdge.Authenticators) {
				authenticatorIndex = 0
			}
			result, err := h.WebApp.PostInput(StateID(r), func() (input interface{}, err error) {
				input = &AuthenticationBeginInput{
					AuthenticatorIndex: authenticatorIndex,
				}
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
