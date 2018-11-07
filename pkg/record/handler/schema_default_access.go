package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	recordGear "github.com/skygeario/skygear-server/pkg/record"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
)

func AttachSchemaDefaultAccessHandler(
	server *server.Server,
	recordDependency recordGear.DependencyMap,
) *server.Server {
	server.Handle("/schema/default_access", &SchemaDefaultAccessHandlerFactory{
		recordDependency,
	}).Methods("POST")
	return server
}

type SchemaDefaultAccessHandlerFactory struct {
	Dependency recordGear.DependencyMap
}

func (f SchemaDefaultAccessHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &SchemaDefaultAccessHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f SchemaDefaultAccessHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireMasterKey),
	)
}

type SchemaDefaultAccessRequestPayload struct {
}

func (s SchemaDefaultAccessRequestPayload) Validate() error {
	return nil
}

/*
SchemaDefaultAccessHandler handles the update of creation access of record
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/schema/default_access <<EOF
{
	"type": "note",
	"default_access": [
		{"public": true, "level": "write"}
	]
}
EOF
*/
type SchemaDefaultAccessHandler struct {
	TxContext   db.TxContext  `dependency:"TxContext"`
	RecordStore record.Store  `dependency:"RecordStore"`
	Logger      *logrus.Entry `dependency:"HandlerLogger"`
}

func (h SchemaDefaultAccessHandler) WithTx() bool {
	return true
}

func (h SchemaDefaultAccessHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := SchemaDefaultAccessRequestPayload{}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (h SchemaDefaultAccessHandler) Handle(req interface{}) (resp interface{}, err error) {
	return
}
