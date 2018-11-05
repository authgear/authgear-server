package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
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

func AttachSchemaCreateHandler(
	server *server.Server,
	recordDependency recordGear.DependencyMap,
) *server.Server {
	server.Handle("/schema/create", &SchemaCreateHandlerFactory{
		recordDependency,
	}).Methods("POST")
	return server
}

type SchemaCreateHandlerFactory struct {
	Dependency recordGear.DependencyMap
}

func (f SchemaCreateHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &SchemaCreateHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f SchemaCreateHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireMasterKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type SchemaCreateRequestPayload struct {
	RawSchemas map[string]schemaFieldList `json:"record_types"`
	Schemas    map[string]record.Schema
}

func (s SchemaCreateRequestPayload) Validate() error {
	for recordType, schema := range s.Schemas {
		if strings.HasPrefix(recordType, "_") {
			return skyerr.NewInvalidArgument("attempts to create reserved table", []string{recordType})
		}
		for fieldName := range schema {
			if strings.HasPrefix(fieldName, "_") {
				return skyerr.NewInvalidArgument("attempts to create reserved field", []string{fieldName})
			}
		}
	}
	return nil
}

/*
SchemaCreateHandler handles the action of creating new columns
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/schema/create <<EOF
{
	"record_types":{
		"student": {
			"fields":[
				{"name": "age", "type": "number"},
				{"name": "nickname", "type": "string"}
			]
		}
	}
}
EOF
*/
type SchemaCreateHandler struct {
	TxContext   db.TxContext  `dependency:"TxContext"`
	RecordStore record.Store  `dependency:"RecordStore"`
	Logger      *logrus.Entry `dependency:"HandlerLogger"`
}

func (h SchemaCreateHandler) WithTx() bool {
	return true
}

func (h SchemaCreateHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := SchemaCreateRequestPayload{}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return nil, err
	}

	payload.Schemas = make(map[string]record.Schema)
	for recordType, schema := range payload.RawSchemas {
		payload.Schemas[recordType] = make(record.Schema)
		for _, field := range schema.Fields {
			var err error
			payload.Schemas[recordType][field.Name], err = record.SimpleNameToFieldType(field.TypeName)
			if err != nil {
				return nil, skyerr.NewInvalidArgument("unexpected field type", []string{field.TypeName})
			}
		}
	}

	return payload, nil
}

func (h SchemaCreateHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(SchemaCreateRequestPayload)

	for recordType, recordSchema := range payload.Schemas {
		_, err = h.RecordStore.Extend(recordType, recordSchema)
		if err != nil {
			h.Logger.WithFields(logrus.Fields{
				"error": err,
				"field": recordType,
			}).Error("fail to extend schema")
			err = skyerr.NewError(skyerr.IncompatibleSchema, err.Error())
			return
		}
	}

	schemas, err := h.RecordStore.GetRecordSchemas()
	if err != nil {
		h.Logger.WithError(err).Error("fail to get record schemas")
		return
	}

	resp = NewSchemaResponse(encodeRecordSchemas(schemas))

	// TODO: send schema change event

	return
}
