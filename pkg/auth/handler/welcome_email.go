package handler

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/welcemail"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
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
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, nil)
}

// ProvideAuthzPolicy provides authorization policy of handler
func (f WelcomeEmailHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireMasterKey),
	)
}

// WelcomeEmailPayload send a dummy welcome email to given email.
//
//  curl -X POST -H "Content-Type: application/json" \
//    -d @- http://localhost:3000/welcome_email/test <<EOF
//  {
//     "email": "xxx@oursky.com",
//     "text_template": "xxx",
//     "html_template": "xxx",
//     "subject": "xxx",
//     "sender": "xxx",
//     "reply_to": "xxx",
//     "sender_name": "xxx",
//     "reply_to_name": "xxx"
//  }
//  EOF
type WelcomeEmailPayload struct {
	Email        string `json:"email"`
	TextTemplate string `json:"text_template"`
	HTMLTemplate string `json:"html_template"`
	Subject      string `json:"subject"`
	Sender       string `json:"sender"`
	ReplyTo      string `json:"reply_to"`
	SenderName   string `json:"sender_name"`
	ReplyToName  string `json:"reply_to_name"`
}

func (payload WelcomeEmailPayload) Validate() error {
	if payload.Email == "" {
		return skyerr.NewInvalidArgument("empty email", []string{"email"})
	}

	return nil
}

// WelcomeEmailHandler handles set disable request
type WelcomeEmailHandler struct {
	WelcomeEmailSender welcemail.TestSender `dependency:"TestWelcomeEmailSender"`
}

func (h WelcomeEmailHandler) WithTx() bool {
	return false
}

// DecodeRequest decode request payload
func (h WelcomeEmailHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := WelcomeEmailPayload{}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return nil, skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}

	return payload, nil
}

// Handle function handle set disabled request
func (h WelcomeEmailHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(WelcomeEmailPayload)
	if err = h.WelcomeEmailSender.Send(
		payload.Email,
		payload.TextTemplate,
		payload.HTMLTemplate,
		payload.Subject,
		payload.Sender,
		payload.ReplyTo,
		payload.SenderName,
		payload.ReplyToName,
	); err == nil {
		resp = "OK"
	}

	return
}
