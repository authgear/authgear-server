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

func AttachSchemaRenameHandler(
	server *server.Server,
	recordDependency recordGear.DependencyMap,
) *server.Server {
	server.Handle("/schema/rename", &SchemaRenameHandlerFactory{
		recordDependency,
	}).Methods("POST")
	return server
}

type SchemaRenameHandlerFactory struct {
	Dependency recordGear.DependencyMap
}

func (f SchemaRenameHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &SchemaRenameHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f SchemaRenameHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireMasterKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type SchemaRenameRequestPayload struct {
}

func (s SchemaRenameRequestPayload) Validate() error {
	return nil
}

/*
SchemaRenameHandler handles the action of renaming column
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/schema/rename <<EOF
{
	"record_type": "student",
	"item_type": "field",
	"item_name": "score",
	"new_name": "exam_score"
}
EOF
*/
type SchemaRenameHandler struct {
	TxContext   db.TxContext  `dependency:"TxContext"`
	RecordStore record.Store  `dependency:"RecordStore"`
	Logger      *logrus.Entry `dependency:"HandlerLogger"`
}

func (h SchemaRenameHandler) WithTx() bool {
	return true
}

func (h SchemaRenameHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := SchemaRenameRequestPayload{}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (h SchemaRenameHandler) Handle(req interface{}) (resp interface{}, err error) {
	return
}
