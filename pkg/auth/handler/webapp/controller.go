package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/auth/webappoauth"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/tester"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type PageService interface {
	UpdateSession(session *webapp.Session) error
	DeleteSession(id string) error
	PeekUncommittedChanges(session *webapp.Session, fn func(graph *interaction.Graph) error) error
	Get(session *webapp.Session) (*interaction.Graph, error)
	GetSession(id string) (*webapp.Session, error)
	GetWithIntent(session *webapp.Session, intent interaction.Intent) (*interaction.Graph, error)
	PostWithIntent(
		session *webapp.Session,
		intent interaction.Intent,
		inputFn func() (interface{}, error),
	) (result *webapp.Result, err error)
	PostWithInput(
		session *webapp.Session,
		inputFn func() (interface{}, error),
	) (result *webapp.Result, err error)
}

type ControllerDeps struct {
	Database                *appdb.Handle
	RedisHandle             *appredis.Handle
	AppID                   config.AppID
	Page                    PageService
	BaseViewModel           *viewmodels.BaseViewModeler
	Renderer                Renderer
	Publisher               *Publisher
	Clock                   clock.Clock
	UIConfig                *config.UIConfig
	ErrorCookie             *webapp.ErrorCookie
	TesterEndpointsProvider tester.EndpointsProvider

	TrustProxy config.TrustProxy
}

type ControllerFactory struct {
	LoggerFactory *log.Factory
	ControllerDeps
}

func (f *ControllerFactory) New(r *http.Request, rw http.ResponseWriter) (*Controller, error) {
	path := r.URL.Path
	ctrl := &Controller{
		Log:            f.LoggerFactory.New(path),
		ControllerDeps: f.ControllerDeps,

		path:         path,
		request:      r,
		response:     rw,
		postHandlers: make(map[string]func() error),
	}

	if err := r.ParseForm(); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return nil, err
	}

	return ctrl, nil
}

type Controller struct {
	Log *log.Logger
	ControllerDeps

	path     string
	request  *http.Request
	response http.ResponseWriter

	getHandler   func() error
	postHandlers map[string]func() error

	skipRewind bool
}

func (c *Controller) RequireUserID() string {
	userID := session.GetUserID(c.request.Context())
	if userID == nil {
		// This should not happen; should have a middleware to redirect user to
		// login page if unauthenticated.
		panic("webapp: require authenticated user")
	}
	return *userID
}

func (c *Controller) RedirectURI() string {
	ruri := webapp.GetRedirectURI(c.request, bool(c.TrustProxy), "")
	// Consent screen will be skipped if redirect uri here is not empty
	// To workaround it we exclude tester endpoint here
	if ruri == c.TesterEndpointsProvider.TesterURL().String() {
		return ""
	}
	return ruri
}

func (c *Controller) Get(fn func() error) {
	c.getHandler = fn
}

func (c *Controller) PostAction(action string, fn func() error) {
	c.postHandlers[action] = fn
}

func (c *Controller) GetSession(id string) (*webapp.Session, error) {
	return c.Page.GetSession(id)
}

func (c *Controller) UpdateSession(s *webapp.Session) error {
	now := c.Clock.NowUTC()
	s.UpdatedAt = now
	err := c.Page.UpdateSession(s)
	if err != nil {
		return err
	}

	msg := &WebsocketMessage{
		Kind: WebsocketMessageKindRefresh,
	}

	err = c.Publisher.Publish(s, msg)
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) DeleteSession(id string) error {
	return c.Page.DeleteSession(id)
}

func (c *Controller) Serve() {
	var err error
	switch c.request.Method {
	case http.MethodGet:
		err = c.Database.WithTx(c.getHandler)
	case http.MethodPost:
		handler, ok := c.postHandlers[c.request.Form.Get("x_action")]
		if !ok {
			http.Error(c.response, "Unknown action", http.StatusBadRequest)
			break
		}

		err = c.Database.WithTx(handler)
	default:
		http.Error(c.response, "Invalid request method", http.StatusMethodNotAllowed)
	}

	if err != nil {
		if apierrors.IsAPIError(err) {
			c.renderError(err)
		} else {
			panic(err)
		}
	}
}

func (c *Controller) renderError(err error) {
	apierror := apierrors.AsAPIError(err)

	// Show WebUIInvalidSession error in different page.
	u := *c.request.URL
	// If the request method is Get, avoid redirect back to the same path
	// which causes infinite redirect loop
	if c.request.Method == http.MethodGet {
		u.Path = "/errors/error"
	}
	if apierror.Reason == webapp.WebUIInvalidSession.Reason {
		u.Path = "/errors/error"
	}

	cookie, err := c.ErrorCookie.SetRecoverableError(c.request, apierror)
	if err != nil {
		panic(err)
	}
	result := webapp.Result{
		RedirectURI:      u.String(),
		NavigationAction: "replace",
		Cookies:          []*http.Cookie{cookie},
	}
	result.WriteResponse(c.response, c.request)
}

func (c *Controller) EntryPointSession(opts webapp.SessionOptions) *webapp.Session {
	s := webapp.GetSession(c.request.Context())
	now := c.Clock.NowUTC()
	o := opts
	if s != nil {
		o = webapp.NewSessionOptionsFromSession(s)
		if opts.RedirectURI != "" {
			o.RedirectURI = opts.RedirectURI
		}
		o.KeepAfterFinish = opts.KeepAfterFinish
	}
	o.UpdatedAt = now
	s = webapp.NewSession(o)
	return s
}

func (c *Controller) EntryPointGet(
	opts webapp.SessionOptions,
	intent interaction.Intent,
) (*interaction.Graph, error) {
	return c.Page.GetWithIntent(c.EntryPointSession(opts), intent)
}

func (c *Controller) EntryPointPost(
	opts webapp.SessionOptions,
	intent interaction.Intent,
	inputFn func() (interface{}, error),
) (*webapp.Result, error) {
	return c.Page.PostWithIntent(c.EntryPointSession(opts), intent, inputFn)
}

func (c *Controller) rewindSessionHistory(session *webapp.Session) error {
	if c.skipRewind {
		// Session has been rewound by caller (alternative step handler),
		// skip rewinding.
		return nil
	}

	stepGraphID := c.request.Form.Get("x_step")
	var curStep *webapp.SessionStep
	for i := len(session.Steps) - 1; i >= 0; i-- {
		step := session.Steps[i]
		if step.GraphID == stepGraphID {
			session.Steps = session.Steps[:i+1]
			curStep = &step
			break
		}
	}

	if curStep == nil && len(session.Steps) > 0 {
		s := session.CurrentStep()
		curStep = &s
	}
	if curStep == nil || !curStep.Kind.MatchPath(c.path) {
		return webapp.ErrSessionStepMismatch
	}
	return nil
}

func (c *Controller) InteractionSession() (*webapp.Session, error) {
	s := webapp.GetSession(c.request.Context())
	if s == nil {
		return nil, webapp.ErrSessionNotFound
	}
	return s, nil
}

func (c *Controller) InteractionGet() (*interaction.Graph, error) {
	s, err := c.InteractionSession()
	if err != nil {
		return nil, err
	}

	if err := c.rewindSessionHistory(s); err != nil {
		return nil, err
	}

	return c.Page.Get(s)
}

func (c *Controller) InteractionGetWithSession(s *webapp.Session) (*interaction.Graph, error) {
	if err := c.rewindSessionHistory(s); err != nil {
		return nil, err
	}

	return c.Page.Get(s)
}

func (c *Controller) InteractionPost(inputFn func() (interface{}, error)) (*webapp.Result, error) {
	s, err := c.InteractionSession()
	if err != nil {
		return nil, err
	}

	if err := c.rewindSessionHistory(s); err != nil {
		return nil, err
	}

	return c.Page.PostWithInput(s, inputFn)
}

func (c *Controller) InteractionOAuthCallback(oauthInput InputOAuthCallback, oauthState *webappoauth.WebappOAuthState) (*webapp.Result, error) {
	inputFn := func() (input interface{}, err error) {
		input = &oauthInput
		return
	}
	// OAuth callback could be triggered by form post from another site (e.g. google)
	// Therefore cookies might not be sent in this case
	// Use the state as web session id
	s, err := c.Page.GetSession(oauthState.WebSessionID)
	if err != nil {
		return nil, err
	}

	if err := c.rewindSessionHistory(s); err != nil {
		return nil, err
	}

	return c.Page.PostWithInput(s, inputFn)
}
