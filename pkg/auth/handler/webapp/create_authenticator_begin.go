package webapp

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureCreateAuthenticatorBeginRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/create_authenticator_begin")
}

type CreateAuthenticatorBeginNode interface {
	GetCreateAuthenticatorEdges() ([]interaction.Edge, error)
	GetCreateAuthenticatorStage() interaction.AuthenticationStage
}

type CreateAuthenticatorBeginHandler struct {
	ControllerFactory ControllerFactory
}

func (h *CreateAuthenticatorBeginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	edgeIndexString := r.Form.Get("x_edge")
	if edgeIndexString == "" {
		edgeIndexString = "0"
	}
	edgeIndex, err := strconv.Atoi(edgeIndexString)
	if err != nil {
		edgeIndex = 0
	}

	ctrl.Get(func() error {
		session, err := ctrl.InteractionSession()
		if err != nil {
			return err
		}

		graph, err := ctrl.InteractionGet()
		if err != nil {
			return err
		}

		var node CreateAuthenticatorBeginNode
		if !graph.FindLastNode(&node) {
			panic("create_authenticator_begin: expected graph has node implementing CreateAuthenticatorBeginNode")
		}
		edges, err := node.GetCreateAuthenticatorEdges()
		if err != nil {
			panic(err)
		}

		if edgeIndex >= len(edges) {
			edgeIndex = 0
		}

		selectedEdge := edges[edgeIndex]
		switch selectedEdge := selectedEdge.(type) {
		case *nodes.EdgeCreateAuthenticatorPassword:
			u := session.CurrentStepURL()
			u.Path = "/create_password"
			http.Redirect(w, r, u.String(), http.StatusFound)
		case *nodes.EdgeCreateAuthenticatorOOBSetup:
			u := session.CurrentStepURL()
			u.Path = "/setup_oob_otp"
			http.Redirect(w, r, u.String(), http.StatusFound)
		case *nodes.EdgeCreateAuthenticatorTOTPSetup:
			result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
				input = &InputSelectTOTP{}
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

type CreateAuthenticatorAlternative struct {
	Type string
	URL  string
}

func DeriveCreateAuthenticatorAlternatives(session *webapp.Session, graph *interaction.Graph, currentType authn.AuthenticatorType) (alternatives []CreateAuthenticatorAlternative, err error) {
	var node CreateAuthenticatorBeginNode
	if !graph.FindLastNode(&node) {
		panic("create_authenticator_begin: expected graph has node implementing CreateAuthenticatorBeginNode")
	}

	var u *url.URL
	for i := len(session.Steps) - 1; i >= 0; i-- {
		step := session.Steps[i]
		if step.Path == "/create_authenticator_begin" {
			u = session.StepURL(i)
			break
		}
	}
	if u == nil {
		panic("authentication_begin: expected session has authentication_begin step")
	}
	q := u.Query()

	edges, err := node.GetCreateAuthenticatorEdges()
	if err != nil {
		return
	}

	for i, edge := range edges {
		q.Set("x_edge", strconv.Itoa(i))
		u.RawQuery = q.Encode()

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
				URL:  u.String(),
			})
		}
	}

	return
}
