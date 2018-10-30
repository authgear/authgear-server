package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/asset"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	recordGear "github.com/skygeario/skygear-server/pkg/record"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
)

func AttachQueryHandler(
	server *server.Server,
	recordDependency recordGear.DependencyMap,
) *server.Server {
	server.Handle("/query", &QueryHandlerFactory{
		recordDependency,
	}).Methods("POST")
	return server
}

type QueryHandlerFactory struct {
	Dependency recordGear.DependencyMap
}

func (f QueryHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &QueryHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f QueryHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type QueryRequestPayload struct {
	Query record.Query
}

func (p QueryRequestPayload) Validate() error {
	return nil
}

/*
QueryHandler is dummy implementation on fetching Records
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/query <<EOF
{
    "record_type": "note",
    "sort": [
        [{"$val": "noteOrder", "$type": "desc"}, "asc"]
    ]
}
EOF
*/
type QueryHandler struct {
	AuthContext auth.ContextGetter `dependency:"AuthContextGetter"`
	TxContext   db.TxContext       `dependency:"TxContext"`
	RecordStore record.Store       `dependency:"RecordStore"`
	Logger      *logrus.Entry      `dependency:"HandlerLogger"`
	AssetStore  asset.Store        `dependency:"AssetStore"`
}

func (h QueryHandler) WithTx() bool {
	return true
}

func (h QueryHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := QueryRequestPayload{}
	parser := QueryParser{UserID: h.AuthContext.AuthInfo().ID}
	data := map[string]interface{}{}
	json.NewDecoder(request.Body).Decode(&data)
	if err := parser.queryFromRaw(data, &payload.Query); err != nil {
		return nil, err
	}

	return payload, nil
}

func (h QueryHandler) Handle(req interface{}) (resp interface{}, err error) {
	return
}
