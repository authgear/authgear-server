package handler

import (
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

func AttachSchemaFetchHandler(
	server *server.Server,
	recordDependency recordGear.DependencyMap,
) *server.Server {
	server.Handle("/schema/fetch", &SchemaFetchHandlerFactory{
		recordDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type SchemaFetchHandlerFactory struct {
	Dependency recordGear.DependencyMap
}

func (f SchemaFetchHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &SchemaFetchHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f SchemaFetchHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireMasterKey),
	)
}

/*
SchemaFetchHandler handles the action of returing information of record schema
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/schema/fetch <<EOF
{
}
EOF
*/
type SchemaFetchHandler struct {
	TxContext   db.TxContext  `dependency:"TxContext"`
	RecordStore record.Store  `dependency:"RecordStore"`
	Logger      *logrus.Entry `dependency:"HandlerLogger"`
}

func (h SchemaFetchHandler) WithTx() bool {
	return true
}

func (h SchemaFetchHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	return handler.EmptyRequestPayload{}, nil
}

func (h SchemaFetchHandler) Handle(req interface{}) (resp interface{}, err error) {
	schemas, err := h.RecordStore.GetRecordSchemas()
	if err != nil {
		h.Logger.WithError(err).Error("fail to get record schemas")
		return
	}

	resp = NewSchemaResponse(encodeRecordSchemas(schemas))

	return
}
