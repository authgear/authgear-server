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

func AttachFieldAccessGetHandler(
	server *server.Server,
	recordDependency recordGear.DependencyMap,
) *server.Server {
	server.Handle("/schema/field_access/get", &FieldAccessGetHandlerFactory{
		recordDependency,
	}).Methods("POST")
	return server
}

type FieldAccessGetHandlerFactory struct {
	Dependency recordGear.DependencyMap
}

func (f FieldAccessGetHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &FieldAccessGetHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f FieldAccessGetHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireMasterKey),
	)
}

/*
FieldAccessGetHandler fetches the entire Field ACL settings.
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/schema/field_access/get <<EOF
{
}
EOF
*/
type FieldAccessGetHandler struct {
	TxContext   db.TxContext  `dependency:"TxContext"`
	RecordStore record.Store  `dependency:"RecordStore"`
	Logger      *logrus.Entry `dependency:"HandlerLogger"`
}

func (h FieldAccessGetHandler) WithTx() bool {
	return true
}

func (h FieldAccessGetHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	return handler.EmptyRequestPayload{}, nil
}

func (h FieldAccessGetHandler) Handle(req interface{}) (resp interface{}, err error) {
	var fieldACL record.FieldACL
	fieldACL, err = h.RecordStore.GetRecordFieldAccess()
	if err != nil {
		h.Logger.WithError(err).Error("fail to get field access")
		return
	}

	resp = NewFieldAccessResponse(fieldACL)

	return
}
