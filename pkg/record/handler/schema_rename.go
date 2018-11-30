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

func AttachSchemaRenameHandler(
	server *server.Server,
	recordDependency recordGear.DependencyMap,
) *server.Server {
	server.Handle("/schema/rename", &SchemaRenameHandlerFactory{
		recordDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type SchemaRenameHandlerFactory struct {
	Dependency recordGear.DependencyMap
}

func (f SchemaRenameHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &SchemaRenameHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f SchemaRenameHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireMasterKey),
	)
}

type SchemaRenameRequestPayload struct {
	RecordType string `json:"record_type"`
	OldName    string `json:"item_name"`
	NewName    string `json:"new_name"`
}

func (p SchemaRenameRequestPayload) Validate() error {
	missingArgs := []string{}
	if p.RecordType == "" {
		missingArgs = append(missingArgs, "record_type")
	}
	if p.OldName == "" {
		missingArgs = append(missingArgs, "item_name")
	}
	if p.NewName == "" {
		missingArgs = append(missingArgs, "new_name")
	}
	if len(missingArgs) > 0 {
		return skyerr.NewInvalidArgument("missing required fields", missingArgs)
	}

	if strings.HasPrefix(p.RecordType, "_") {
		return skyerr.NewInvalidArgument("attempts to change reserved table", []string{"record_type"})
	}
	if strings.HasPrefix(p.OldName, "_") {
		return skyerr.NewInvalidArgument("attempts to change reserved key", []string{"item_name"})
	}
	if strings.HasPrefix(p.NewName, "_") {
		return skyerr.NewInvalidArgument("attempts to change reserved key", []string{"new_name"})
	}
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
	payload := req.(SchemaRenameRequestPayload)

	if err = h.RecordStore.RenameSchema(payload.RecordType, payload.OldName, payload.NewName); err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error": err,
			"field": payload.RecordType,
		}).Error("fail to rename schema")
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
