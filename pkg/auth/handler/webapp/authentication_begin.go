package webapp

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/core/authn"
	corephone "github.com/authgear/authgear-server/pkg/core/phone"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/mail"
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

type AuthenticationAlternative struct {
	Type         string
	URL          string
	MaskedTarget string
}

func DeriveAuthenticationAlternatives(stateID string, graph *newinteraction.Graph, currentType authn.AuthenticatorType, currentTarget string) (alternatives []AuthenticationAlternative) {
	var node AuthenticationBeginNode
	if !graph.FindLastNode(&node) {
		panic("authentication_begin: expected graph has node implementing AuthenticationBeginNode")
	}

	edges := node.GetAuthenticationEdges()

	for i, edge := range edges {
		switch edge := edge.(type) {
		case *nodes.EdgeAuthenticationPassword:
			typ := authn.AuthenticatorTypePassword
			if typ != currentType {
				q := url.Values{}
				q.Set("x_edge", strconv.Itoa(i))
				alternatives = append(alternatives, AuthenticationAlternative{
					Type: string(typ),
					URL: webapp.AttachStateID(stateID, &url.URL{
						Path:     "/authentication_begin",
						RawQuery: q.Encode(),
					}).String(),
				})
			}
		case *nodes.EdgeAuthenticationTOTP:
			typ := authn.AuthenticatorTypeTOTP
			if typ != currentType {
				q := url.Values{}
				q.Set("x_edge", strconv.Itoa(i))
				alternatives = append(alternatives, AuthenticationAlternative{
					Type: string(typ),
					URL: webapp.AttachStateID(stateID, &url.URL{
						Path:     "/authentication_begin",
						RawQuery: q.Encode(),
					}).String(),
				})
			}
		case *nodes.EdgeAuthenticationOOBTrigger:
			typ := authn.AuthenticatorTypeOOB
			if typ != currentType {
				for j, a := range edge.Authenticators {
					channel := a.Props[authenticator.AuthenticatorPropOOBOTPChannelType].(string)

					var target string
					var maskedTarget string
					switch channel {
					case string(authn.AuthenticatorOOBChannelSMS):
						phone := a.Props[authenticator.AuthenticatorPropOOBOTPPhone].(string)
						target = phone
						maskedTarget = corephone.Mask(phone)
					case string(authn.AuthenticatorOOBChannelEmail):
						email := a.Props[authenticator.AuthenticatorPropOOBOTPEmail].(string)
						target = email
						maskedTarget = mail.MaskAddress(email)
					default:
						panic("authentication_begin: unexpected channel: " + channel)
					}

					if currentTarget == target {
						continue
					}

					q := url.Values{}
					q.Set("x_edge", strconv.Itoa(i))
					q.Set("x_authenticator", strconv.Itoa(j))
					alternatives = append(alternatives, AuthenticationAlternative{
						Type: string(typ),
						URL: webapp.AttachStateID(stateID, &url.URL{
							Path:     "/authentication_begin",
							RawQuery: q.Encode(),
						}).String(),
						MaskedTarget: maskedTarget,
					})
				}
			}
		default:
			panic(fmt.Errorf("authentication_begin: unexpected edge: %T", edge))
		}
	}

	return
}
