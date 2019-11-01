package handler

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/welcemail"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

// AttachWelcomeEmailHandler attaches WelcomeEmailHandler to server
func AttachWelcomeEmailHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/welcome_email/test", &WelcomeEmailHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

// WelcomeEmailHandlerFactory creates WelcomeEmailHandler
type WelcomeEmailHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new WelcomeEmailHandler
func (f WelcomeEmailHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &WelcomeEmailHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
}

type WelcomeEmailPayload struct {
	Email        string `json:"email"`
	TextTemplate string `json:"text_template"`
	HTMLTemplate string `json:"html_template"`
	Subject      string `json:"subject"`
	Sender       string `json:"sender"`
	ReplyTo      string `json:"reply_to"`
}

const WelcomeEmailTestRequestSchema = `
{
	"$id": "#WelcomeEmailTestRequest",
	"type": "object",
	"properties": {
		"email": { "type": "string", "format": "email" },
		"text_template": { "type": "string", "minLength": 1 },
		"html_template": { "type": "string", "minLength": 1 },
		"subject": { "type": "string", "minLength": 1 },
		"sender": { "type": "string", "minLength": 1 },
		"reply_to": { "type": "string", "minLength": 1 }
	}
}
`

// WelcomeEmailHandler send a dummy welcome email to given email.
//
//  curl -X POST -H "Content-Type: application/json" \
//    -d @- http://localhost:3000/welcome_email/test <<EOF
//  {
//     "email": "xxx@oursky.com",
//     "text_template": "xxx",
//     "html_template": "xxx",
//     "subject": "xxx",
//     "sender": "xxx",
//     "reply_to": "xxx"
//  }
//  EOF
type WelcomeEmailHandler struct {
	RequireAuthz       handler.RequireAuthz  `dependency:"RequireAuthz"`
	Validator          *validation.Validator `dependency:"Validator"`
	WelcomeEmailSender welcemail.TestSender  `dependency:"TestWelcomeEmailSender"`
}

// ProvideAuthzPolicy provides authorization policy of handler
func (h WelcomeEmailHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireMasterKey),
	)
}

func (h WelcomeEmailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h WelcomeEmailHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	var payload WelcomeEmailPayload
	if err := handler.BindJSONBody(r, w, h.Validator, "#WelcomeEmailTestRequest", &payload); err != nil {
		return nil, err
	}
	if err = h.WelcomeEmailSender.Send(
		payload.Email,
		payload.TextTemplate,
		payload.HTMLTemplate,
		payload.Subject,
		payload.Sender,
		payload.ReplyTo,
	); err == nil {
		resp = struct{}{}
	}

	return
}
