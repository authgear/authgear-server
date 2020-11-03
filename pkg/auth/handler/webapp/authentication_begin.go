package webapp

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	corephone "github.com/authgear/authgear-server/pkg/util/phone"
)

func ConfigureAuthenticationBeginRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/authentication_begin")
}

type AuthenticationBeginNode interface {
	GetAuthenticationEdges() ([]interaction.Edge, error)
}

type AuthenticationBeginHandler struct {
	Database             *db.Handle
	WebApp               WebAppService
	MFADeviceTokenCookie mfa.CookieDef
}

func (h *AuthenticationBeginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var err error

	var firstTimeEnterHere bool
	deviceTokenCookie, _ := r.Cookie(h.MFADeviceTokenCookie.Def.Name)

	edgeIndexString := r.Form.Get("x_edge")
	if edgeIndexString == "" {
		firstTimeEnterHere = true
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
	var graph *interaction.Graph

	err = h.Database.WithTx(func() error {
		state, graph, err = h.WebApp.Get(StateID(r))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	var node AuthenticationBeginNode
	if !graph.FindLastNode(&node) {
		panic("authentication_begin: expected graph has node implementing AuthenticationBeginNode")
	}
	edges, err := node.GetAuthenticationEdges()
	if err != nil {
		panic(err)
	}
	if edgeIndex >= len(edges) {
		edgeIndex = 0
	}

	err = h.Database.WithTx(func() error {
		if firstTimeEnterHere && deviceTokenCookie != nil {
			for _, edge := range edges {
				if _, ok := edge.(*nodes.EdgeUseDeviceToken); ok {
					result, err := h.WebApp.PostInput(StateID(r), func() (input interface{}, err error) {
						input = &InputAuthDeviceToken{
							DeviceToken: deviceTokenCookie.Value,
						}
						return
					})
					if err != nil {
						return err
					}

					result.WriteResponse(w, r)
					return nil
				}
			}
		}

		selectedEdge := edges[edgeIndex]
		switch selectedEdge := selectedEdge.(type) {
		case *nodes.EdgeConsumeRecoveryCode:
			http.Redirect(w, r, webapp.AttachStateID(state.ID, &url.URL{
				Path: "/enter_recovery_code",
			}).String(), http.StatusFound)
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
				input = &InputTriggerOOB{
					AuthenticatorIndex: authenticatorIndex,
				}
				return
			})
			if err != nil {
				return err
			}
			result.WriteResponse(w, r)
		default:
			panic(fmt.Errorf("webapp: unexpected edge: %T", selectedEdge))
		}

		return nil
	})
	if err != nil {
		panic(err)
	}
}

type AuthenticationType string

const (
	AuthenticationTypePassword                        = AuthenticationType(string(authn.AuthenticatorTypePassword))
	AuthenticationTypeTOTP                            = AuthenticationType(string(authn.AuthenticatorTypeTOTP))
	AuthenticationTypeOOB                             = AuthenticationType(string(authn.AuthenticatorTypeOOB))
	AuthenticationTypeRecoveryCode AuthenticationType = "recovery_code"
	AuthenticationTypeDeviceToken  AuthenticationType = "device_token"
)

type AuthenticationAlternative struct {
	Type         string
	URL          string
	MaskedTarget string
}

func DeriveAuthenticationAlternatives(stateID string, graph *interaction.Graph, currentType AuthenticationType, currentTarget string) (alternatives []AuthenticationAlternative, err error) {
	var node AuthenticationBeginNode
	if !graph.FindLastNode(&node) {
		panic("authentication_begin: expected graph has node implementing AuthenticationBeginNode")
	}

	edges, err := node.GetAuthenticationEdges()
	if err != nil {
		return nil, err
	}

	for i, edge := range edges {
		switch edge := edge.(type) {
		case *nodes.EdgeUseDeviceToken:
			alternatives = append(alternatives, AuthenticationAlternative{
				Type: string(AuthenticationTypeDeviceToken),
			})
		case *nodes.EdgeConsumeRecoveryCode:
			typ := AuthenticationTypeRecoveryCode
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
		case *nodes.EdgeAuthenticationPassword:
			typ := AuthenticationTypePassword
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
			typ := AuthenticationTypeTOTP
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
			typ := AuthenticationTypeOOB
			if typ != currentType {
				for j, a := range edge.Authenticators {
					channel := a.Claims[authenticator.AuthenticatorClaimOOBOTPChannelType].(string)

					var target string
					var maskedTarget string
					switch channel {
					case string(authn.AuthenticatorOOBChannelSMS):
						phone := a.Claims[authenticator.AuthenticatorClaimOOBOTPPhone].(string)
						target = phone
						maskedTarget = corephone.Mask(phone)
					case string(authn.AuthenticatorOOBChannelEmail):
						email := a.Claims[authenticator.AuthenticatorClaimOOBOTPEmail].(string)
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
