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

func AttachDeleteHandler(
	server *server.Server,
	recordDependency recordGear.DependencyMap,
) *server.Server {
	server.Handle("/delete", &DeleteHandlerFactory{
		recordDependency,
	}).Methods("POST")
	return server
}

type DeleteHandlerFactory struct {
	Dependency recordGear.DependencyMap
}

func (f DeleteHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &DeleteHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f DeleteHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type DeleteRequestPayload struct {
}

func (p DeleteRequestPayload) Validate() error {
	return nil
}

/*
DeleteHandler is dummy implementation on delete Records
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/delete <<EOF
{
    "records": [
        {
            "_recordType": "note",
            "_recordID": "EA6A3E68-90F3-49B5-B470-5FFDB7A0D4E8"
        }
    ]
}
EOF

Deprecated format:
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/delete <<EOF
{
    "ids": ["note/EA6A3E68-90F3-49B5-B470-5FFDB7A0D4E8"]
}
EOF
*/
type DeleteHandler struct {
	AuthContext auth.ContextGetter `dependency:"AuthContextGetter"`
	TxContext   db.TxContext       `dependency:"TxContext"`
	RecordStore record.Store       `dependency:"RecordStore"`
	Logger      *logrus.Entry      `dependency:"HandlerLogger"`
	AssetStore  asset.Store        `dependency:"AssetStore"`
}

func (h DeleteHandler) WithTx() bool {
	return true
}

func (h DeleteHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := DeleteRequestPayload{}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (h DeleteHandler) Handle(req interface{}) (resp interface{}, err error) {
	return
}
