package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
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

type RecordDeleteRecordPayload struct {
	Type string `json:"_recordType"`
	Key  string `json:"_recordID"`
}

func (p RecordDeleteRecordPayload) RecordID() record.ID {
	return record.ID{
		Type: p.Type,
		Key:  p.Key,
	}
}

type DeleteRequestPayload struct {
	DeprecatedIDs   []string                    `json:"ids"`
	RawRecords      []RecordDeleteRecordPayload `json:"records"`
	Atomic          bool                        `json:"atomic"`
	parsedRecordIDs []record.ID
}

func (p DeleteRequestPayload) Validate() error {
	if len(p.parsedRecordIDs) == 0 {
		return skyerr.NewInvalidArgument("expected list of records", []string{"records"})
	}

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

	if len(payload.RawRecords) > 0 {
		length := len(payload.RawRecords)
		payload.parsedRecordIDs = make([]record.ID, length, length)
		for i, rawRecord := range payload.RawRecords {
			payload.parsedRecordIDs[i] = rawRecord.RecordID()
		}
	} else if len(payload.DeprecatedIDs) > 0 {
		// NOTE(cheungpat): Handling for deprecated fields.
		length := len(payload.DeprecatedIDs)
		payload.parsedRecordIDs = make([]record.ID, length, length)
		for i, rawID := range payload.DeprecatedIDs {
			ss := strings.SplitN(rawID, "/", 2)
			if len(ss) == 1 {
				return nil, skyerr.NewInvalidArgument(
					`record: "_id" should be of format '{type}/{id}', got "`+rawID+`"`,
					[]string{"ids"},
				)
			}

			payload.parsedRecordIDs[i].Type = ss[0]
			payload.parsedRecordIDs[i].Key = ss[1]
		}
	}

	return payload, nil
}

func (h DeleteHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(DeleteRequestPayload)
	fmt.Printf("%+v\n", payload)
	return
}
