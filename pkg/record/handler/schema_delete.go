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

func AttachSchemaDeleteHandler(
	server *server.Server,
	recordDependency recordGear.DependencyMap,
) *server.Server {
	server.Handle("/schema/delete", &SchemaDeleteHandlerFactory{
		recordDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type SchemaDeleteHandlerFactory struct {
	Dependency recordGear.DependencyMap
}

func (f SchemaDeleteHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &SchemaDeleteHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f SchemaDeleteHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireMasterKey),
	)
}

type SchemaDeleteRequestPayload struct {
	RecordType string `json:"record_type"`
	ColumnName string `json:"item_name"`
}

func (p SchemaDeleteRequestPayload) Validate() error {
	missingArgs := []string{}
	if p.RecordType == "" {
		missingArgs = append(missingArgs, "record_type")
	}
	if p.ColumnName == "" {
		missingArgs = append(missingArgs, "item_name")
	}
	if len(missingArgs) > 0 {
		return skyerr.NewInvalidArgument("missing required fields", missingArgs)
	}

	if strings.HasPrefix(p.RecordType, "_") {
		return skyerr.NewInvalidArgument("attempts to change reserved table", []string{"record_type"})
	}

	if strings.HasPrefix(p.ColumnName, "_") {
		return skyerr.NewInvalidArgument("attempts to change reserved key", []string{"item_name"})
	}
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
	payload := req.(SchemaDeleteRequestPayload)

	if err = h.RecordStore.DeleteSchema(payload.RecordType, payload.ColumnName); err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error": err,
			"field": payload.RecordType,
		}).Error("fail to delete schema")
		err = skyerr.NewError(skyerr.ResourceNotFound, err.Error())
		return
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
