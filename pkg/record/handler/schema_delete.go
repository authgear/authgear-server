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

func AttachSchemaDeleteHandler(
	server *server.Server,
	recordDependency recordGear.DependencyMap,
) *server.Server {
	server.Handle("/schema/delete", &SchemaDeleteHandlerFactory{
		recordDependency,
	}).Methods("POST")
	return server
}

type SchemaDeleteHandlerFactory struct {
	Dependency recordGear.DependencyMap
}

func (f SchemaDeleteHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &SchemaDeleteHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f SchemaDeleteHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireMasterKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type SchemaDeleteRequestPayload struct {
}

func (s SchemaDeleteRequestPayload) Validate() error {
	return nil
}

/*
SchemaDeleteHandler handles the action of deleting column
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/schema/delete <<EOF
{
	"record_type": "student",
	"item_name": "score"
}
EOF
*/
type SchemaDeleteHandler struct {
	TxContext   db.TxContext  `dependency:"TxContext"`
	RecordStore record.Store  `dependency:"RecordStore"`
	Logger      *logrus.Entry `dependency:"HandlerLogger"`
}

func (h SchemaDeleteHandler) WithTx() bool {
	return true
}

func (h SchemaDeleteHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := SchemaDeleteRequestPayload{}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (h SchemaDeleteHandler) Handle(req interface{}) (resp interface{}, err error) {
	return
}
