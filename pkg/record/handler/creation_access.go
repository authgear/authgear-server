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

func AttachCreationAccessHandler(
	server *server.Server,
	recordDependency recordGear.DependencyMap,
) *server.Server {
	server.Handle("/schema/access", &CreationAccessHandlerFactory{
		recordDependency,
	}).Methods("POST")
	return server
}

type CreationAccessHandlerFactory struct {
	Dependency recordGear.DependencyMap
}

func (f CreationAccessHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &CreationAccessHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f CreationAccessHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireMasterKey),
	)
}

type CreationAccessRequestPayload struct {
}

func (s CreationAccessRequestPayload) Validate() error {
	return nil
}

/*
CreationAccessHandler handles the update of creation access of record
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/schema/access <<EOF
{
	"type": "note",
	"create_roles": [
		"admin",
		"writer"
	]
}
EOF
*/
type CreationAccessHandler struct {
	TxContext   db.TxContext  `dependency:"TxContext"`
	RecordStore record.Store  `dependency:"RecordStore"`
	Logger      *logrus.Entry `dependency:"HandlerLogger"`
}

func (h CreationAccessHandler) WithTx() bool {
	return true
}

func (h CreationAccessHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := CreationAccessRequestPayload{}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (h CreationAccessHandler) Handle(req interface{}) (resp interface{}, err error) {
	return
}
