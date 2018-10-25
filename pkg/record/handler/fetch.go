package handler

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	recordGear "github.com/skygeario/skygear-server/pkg/record"
)

func AttachFetchHandler(
	server *server.Server,
	recordDependency recordGear.DependencyMap,
) *server.Server {
	server.Handle("/fetch", &FetchHandlerFactory{
		recordDependency,
	}).Methods("POST")
	return server
}

type FetchHandlerFactory struct {
	Dependency recordGear.DependencyMap
}

func (f FetchHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &FetchHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f FetchHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type FetchRequestPayload struct {
}

func (s FetchRequestPayload) Validate() error {
	return nil
}

/*
FetchHandler is dummy implementation on fetching Records
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/fetch <<EOF
{
    "ids": ["note/1004", "note/1005"]
}
EOF
*/
type FetchHandler struct {
	TxContext db.TxContext `dependency:"TxContext"`
}

func (h FetchHandler) WithTx() bool {
	return true
}

func (h FetchHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := FetchRequestPayload{}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (h FetchHandler) Handle(req interface{}) (resp interface{}, err error) {
	return
}
