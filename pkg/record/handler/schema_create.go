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

func AttachSchemaCreateHandler(
	server *server.Server,
	recordDependency recordGear.DependencyMap,
) *server.Server {
	server.Handle("/schema/create", &SchemaCreateHandlerFactory{
		recordDependency,
	}).Methods("POST")
	return server
}

type SchemaCreateHandlerFactory struct {
	Dependency recordGear.DependencyMap
}

func (f SchemaCreateHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &SchemaCreateHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f SchemaCreateHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireMasterKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type SchemaCreateRequestPayload struct {
}

func (s SchemaCreateRequestPayload) Validate() error {
	return nil
}

/*
SchemaCreateHandler handles the action of creating new columns
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/schema/create <<EOF
{
	"record_types":{
		"student": {
			"fields":[
				{"name": "age", "type": "number"},
				{"name": "nickname" "type": "string"}
			]
		}
	}
}
EOF
*/
type SchemaCreateHandler struct {
	TxContext db.TxContext `dependency:"TxContext"`
}

func (h SchemaCreateHandler) WithTx() bool {
	return true
}

func (h SchemaCreateHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := SchemaCreateRequestPayload{}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (h SchemaCreateHandler) Handle(req interface{}) (resp interface{}, err error) {
	return
}
