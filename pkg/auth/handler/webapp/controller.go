package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type PageService interface {
	UpdateSession(session *webapp.Session) error
	Get(path string, session *webapp.Session) (*interaction.Graph, error)
	GetWithIntent(session *webapp.Session, intent interaction.Intent) (*interaction.Graph, error)
	PostWithIntent(
		session *webapp.Session,
		intent interaction.Intent,
		inputFn func() (interface{}, error),
	) (result *webapp.Result, err error)
	PostWithInput(
		path string,
		session *webapp.Session,
		inputFn func() (interface{}, error),
	) (result *webapp.Result, err error)
}

type ControllerDeps struct {
	Database *db.Handle
	Page     PageService
	Renderer Renderer

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

		path:     path,
		request:  r,
		response: rw,
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
	return webapp.GetRedirectURI(c.request, bool(c.TrustProxy), "")
}

func (c *Controller) Get(fn func() error) {
	if c.request.Method != "GET" {
		return
	}

	err := c.Database.WithTx(func() error {
		return fn()
	})
	if err != nil {
		panic(err)
	}
}

func (c *Controller) PostAction(action string, fn func() error) {
	if c.request.Method != "POST" || c.request.Form.Get("x_action") != action {
		return
	}

	err := c.Database.WithTx(func() error {
		return fn()
	})
	if err != nil {
		panic(err)
	}
}

func (c *Controller) EntryPointSession(opts webapp.SessionOptions) *webapp.Session {
	s := webapp.GetSession(c.request.Context())
	if s == nil {
		s = webapp.NewSession(opts)
	} else {
		s = webapp.NewSession(webapp.NewSessionOptionsFromSession(s))
		if opts.RedirectURI != "" {
			s.RedirectURI = opts.RedirectURI
		}
		s.KeepAfterFinish = opts.KeepAfterFinish
	}

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

	return c.Page.Get(c.path, s)
}

func (c *Controller) InteractionPost(inputFn func() (interface{}, error)) (*webapp.Result, error) {
	s, err := c.InteractionSession()
	if err != nil {
		return nil, err
	}

	return c.Page.PostWithInput(c.path, s, inputFn)
}
