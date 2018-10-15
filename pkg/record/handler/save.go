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
	"github.com/skygeario/skygear-server/pkg/record"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/skyconv"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func AttachSaveHandler(
	server *server.Server,
	recordDependency record.DependencyMap,
) *server.Server {
	server.Handle("/save", &RecordHandlerFactory{
		recordDependency,
	}).Methods("POST")
	return server
}

type RecordHandlerFactory struct {
	Dependency record.DependencyMap
}

func (f RecordHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &SaveHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f RecordHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type SaveRequestPayload struct {
	Atomic bool `json:"atomic"`

	// RawMaps stores the original incoming `records`.
	RawMaps []map[string]interface{} `json:"records"`

	// IncomigItems contains de-serialized recordID or de-serialization error,
	// the item is one-one corresponding to RawMaps.
	IncomingItems []interface{}

	// Records contains the successfully de-serialized record
	Records []*skydb.Record

	// Errs is the array of de-serialization errors
	Errs []skyerr.Error
}

func (s SaveRequestPayload) Validate() error {
	if len(s.RawMaps) == 0 {
		return skyerr.NewInvalidArgument("expected list of record", []string{"records"})
	}

	return nil
}

func (s SaveRequestPayload) isClean() bool {
	return len(s.Errs) == 0
}

/*
SaveHandler is dummy implementation on save/modify Records
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/save <<EOF
{
    "records": [{
        "_id": "note/EA6A3E68-90F3-49B5-B470-5FFDB7A0D4E8",
        "content": "ewdsa",
        "_access": [{
            "role": "admin",
            "level": "write"
        }]
    }]
}
EOF

Save with reference
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/save <<EOF
{
  "records": [
    {
      "collection": {
        "$type": "ref",
        "$id": "collection/10"
      },
      "noteOrder": 1,
      "content": "hi",
      "_id": "note/71BAE736-E9C5-43CB-ADD1-D8633B80CAFA",
      "_type": "record",
      "_access": [{
          "role": "admin",
          "level": "write"
      }]
    }
  ]
}
EOF
*/
type SaveHandler struct {
	TxContext db.TxContext `dependency:"TxContext"`
}

func (h SaveHandler) WithTx() bool {
	return false
}

func (h SaveHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := SaveRequestPayload{}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return nil, err
	}

	for _, recordMap := range payload.RawMaps {
		var record skydb.Record
		if err := (*skyconv.JSONRecord)(&record).FromMap(recordMap); err != nil {
			skyErr := skyerr.NewError(skyerr.InvalidArgument, err.Error())
			payload.Errs = append(payload.Errs, skyErr)
			payload.IncomingItems = append(payload.IncomingItems, skyErr)
		} else {
			record.SanitizeForInput()
			payload.IncomingItems = append(payload.IncomingItems, record.ID)
			payload.Records = append(payload.Records, &record)
		}
	}

	return payload, nil
}

func (h SaveHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(SaveRequestPayload)

	// TODO: Implement record save handler
	resp = payload

	return
}
