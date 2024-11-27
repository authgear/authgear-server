package webapp

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/tester"
	"github.com/authgear/authgear-server/pkg/lib/webappoauth"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type PageService interface {
	UpdateSession(ctx context.Context, session *webapp.Session) error
	DeleteSession(ctx context.Context, id string) error
	PeekUncommittedChanges(ctx context.Context, session *webapp.Session, fn func(graph *interaction.Graph) error) error
	Get(ctx context.Context, session *webapp.Session) (*interaction.Graph, error)
	GetSession(ctx context.Context, id string) (*webapp.Session, error)
	GetWithIntent(ctx context.Context, session *webapp.Session, intent interaction.Intent) (*interaction.Graph, error)
	PostWithIntent(
		ctx context.Context,
		session *webapp.Session,
		intent interaction.Intent,
		inputFn func() (interface{}, error),
	) (result *webapp.Result, err error)
	PostWithInput(
		ctx context.Context,
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
	TesterEndpointsProvider tester.EndpointsProvider
	ErrorRenderer           *ErrorRenderer

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
		postHandlers: make(map[string]func(ctx context.Context) error),
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

	// preHandler will be executed before any handler executed
	preHandler   func(ctx context.Context) error
	getHandler   func(ctx context.Context) error
	postHandlers map[string]func(ctx context.Context) error

	skipRewind bool
}

func (c *Controller) RequireUserID(ctx context.Context) string {
	userID := session.GetUserID(ctx)
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

func (c *Controller) Get(fn func(ctx context.Context) error) {
	c.getHandler = fn
}

func (c *Controller) BeforeHandle(fn func(ctx context.Context) error) {
	c.preHandler = fn
}

func (c *Controller) PostAction(action string, fn func(ctx context.Context) error) {
	c.postHandlers[action] = fn
}

func (c *Controller) GetSession(ctx context.Context, id string) (*webapp.Session, error) {
	return c.Page.GetSession(ctx, id)
}

func (c *Controller) UpdateSession(ctx context.Context, s *webapp.Session) error {
	now := c.Clock.NowUTC()
	s.UpdatedAt = now
	err := c.Page.UpdateSession(ctx, s)
	if err != nil {
		return err
	}

	msg := &WebsocketMessage{
		Kind: WebsocketMessageKindRefresh,
	}

	err = c.Publisher.Publish(ctx, s, msg)
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) DeleteSession(ctx context.Context, id string) error {
	return c.Page.DeleteSession(ctx, id)
}

func (c *Controller) ServeWithoutDBTx(ctx context.Context) {
	c.serve(ctx, serveOption{
		withDBTx: false,
	})
}

func (c *Controller) ServeWithDBTx(ctx context.Context) {
	c.serve(ctx, serveOption{
		withDBTx: true,
	})
}

type serveOption struct {
	withDBTx bool
}

func (c *Controller) serve(ctx context.Context, option serveOption) {
	var err error
	fns := [](func(ctx context.Context) error){
		func(ctx context.Context) error {
			if c.preHandler != nil {
				return c.preHandler(ctx)
			}
			return nil
		},
	}
	switch c.request.Method {
	case http.MethodGet:
		fns = append(fns, c.getHandler)
	case http.MethodPost:
		handler, ok := c.postHandlers[c.request.Form.Get("x_action")]
		if !ok {
			http.Error(c.response, "Unknown action", http.StatusBadRequest)
			break
		}

		fns = append(fns, handler)
	default:
		http.Error(c.response, "Invalid request method", http.StatusMethodNotAllowed)
	}

	handleFunc := func(ctx context.Context) error {
		for _, fn := range fns {
			err := fn(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if option.withDBTx {
		err = c.Database.WithTx(ctx, handleFunc)
	} else {
		err = handleFunc(ctx)
	}

	if err != nil {
		if apierrors.IsAPIError(err) {
			c.renderError(ctx, err)
		} else {
			panic(err)
		}
	}
}

func (c *Controller) renderError(ctx context.Context, err error) {
	c.ErrorRenderer.RenderError(ctx, c.response, c.request, err)
}

func (c *Controller) EntryPointSession(ctx context.Context, opts webapp.SessionOptions) *webapp.Session {
	s := webapp.GetSession(ctx)
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
	ctx context.Context,
	opts webapp.SessionOptions,
	intent interaction.Intent,
) (*interaction.Graph, error) {
	return c.Page.GetWithIntent(ctx, c.EntryPointSession(ctx, opts), intent)
}

func (c *Controller) EntryPointPost(
	ctx context.Context,
	opts webapp.SessionOptions,
	intent interaction.Intent,
	inputFn func() (interface{}, error),
) (*webapp.Result, error) {
	return c.Page.PostWithIntent(ctx, c.EntryPointSession(ctx, opts), intent, inputFn)
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

func (c *Controller) InteractionSession(ctx context.Context) (*webapp.Session, error) {
	s := webapp.GetSession(ctx)
	if s == nil {
		return nil, webapp.ErrSessionNotFound
	}
	return s, nil
}

func (c *Controller) InteractionGet(ctx context.Context) (*interaction.Graph, error) {
	s, err := c.InteractionSession(ctx)
	if err != nil {
		return nil, err
	}

	if err := c.rewindSessionHistory(s); err != nil {
		return nil, err
	}

	return c.Page.Get(ctx, s)
}

func (c *Controller) InteractionGetWithSession(ctx context.Context, s *webapp.Session) (*interaction.Graph, error) {
	if err := c.rewindSessionHistory(s); err != nil {
		return nil, err
	}

	return c.Page.Get(ctx, s)
}

func (c *Controller) InteractionPost(ctx context.Context, inputFn func() (interface{}, error)) (*webapp.Result, error) {
	s, err := c.InteractionSession(ctx)
	if err != nil {
		return nil, err
	}

	if err := c.rewindSessionHistory(s); err != nil {
		return nil, err
	}

	return c.Page.PostWithInput(ctx, s, inputFn)
}

func (c *Controller) InteractionOAuthCallback(ctx context.Context, oauthInput InputOAuthCallback, oauthState *webappoauth.WebappOAuthState) (*webapp.Result, error) {
	inputFn := func() (input interface{}, err error) {
		input = &oauthInput
		return
	}
	// OAuth callback could be triggered by form post from another site (e.g. google)
	// Therefore cookies might not be sent in this case
	// Use the state as web session id
	s, err := c.Page.GetSession(ctx, oauthState.WebSessionID)
	if err != nil {
		return nil, err
	}

	if err := c.rewindSessionHistory(s); err != nil {
		return nil, err
	}

	return c.Page.PostWithInput(ctx, s, inputFn)
}
