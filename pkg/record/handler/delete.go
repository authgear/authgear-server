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
	"github.com/skygeario/skygear-server/pkg/core/model"
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
	RawRecords      []RecordDeleteRecordPayload `json:"records"`
	Atomic          bool                        `json:"atomic"`
	parsedRecordIDs []record.ID
}

func (p DeleteRequestPayload) Validate() error {
	if len(p.parsedRecordIDs) == 0 {
		return skyerr.NewInvalidArgument("expected list of records", []string{"records"})
	}

	for _, id := range p.parsedRecordIDs {
		if id.Type == "" {
			return skyerr.NewInvalidArgument("expected record type", []string{"_recordType"})
		}

		if id.Key == "" {
			return skyerr.NewInvalidArgument("expected record id", []string{"_recordID"})
		}
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
*/
type DeleteHandler struct {
	AuthContext auth.ContextGetter `dependency:"AuthContextGetter"`
	TxContext   db.TxContext       `dependency:"TxContext"`
	RecordStore record.Store       `dependency:"RecordStore"`
	Logger      *logrus.Entry      `dependency:"HandlerLogger"`
	AssetStore  asset.Store        `dependency:"AssetStore"`
}

func (h DeleteHandler) WithTx() bool {
	return false
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
	}

	return payload, nil
}

func (h DeleteHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(DeleteRequestPayload)

	modifyReq := RecordModifyRequest{
		RecordStore:       h.RecordStore,
		TxContext:         h.TxContext,
		AssetStore:        h.AssetStore,
		Logger:            h.Logger,
		AuthInfo:          h.AuthContext.AuthInfo(),
		RecordIDsToDelete: payload.parsedRecordIDs,
		Atomic:            payload.Atomic,
		WithMasterKey:     h.AuthContext.AccessKeyType() == model.MasterAccessKey,
	}
	modifyResp := RecordModifyResponse{
		ErrMap: map[record.ID]skyerr.Error{},
	}

	// Open transaction for whole operation if atomic save
	if payload.Atomic {
		if err = h.TxContext.BeginTx(); err != nil {
			return
		}
	}

	var opErr error
	defer func() {
		if payload.Atomic {
			if txErr := db.EndTx(h.TxContext, opErr); txErr != nil {
				err = txErr
			}
		} else {
			err = opErr
		}
	}()

	if opErr = RecordDeleteHandler(&modifyReq, &modifyResp); opErr != nil {
		h.Logger.WithError(err).Debugf("Failed to delete records")

		// Override error in atomic save
		if payload.Atomic && len(modifyResp.ErrMap) > 0 {
			info := map[string]interface{}{}
			for recordID, err := range modifyResp.ErrMap {
				info[recordID.String()] = err
			}

			opErr = skyerr.NewErrorWithInfo(skyerr.AtomicOperationFailure,
				"Atomic Operation rolled back due to one or more errors",
				info)
			return
		}

		return
	}

	results := make([]interface{}, 0, len(payload.parsedRecordIDs))
	h.makeResultsForRecordIDs(payload.parsedRecordIDs, modifyResp, &results)

	resp = results

	return
}

func (h DeleteHandler) makeResultsForRecordIDs(recordIDs []record.ID, resp RecordModifyResponse, results *[]interface{}) {
	for _, recordID := range recordIDs {
		var result interface{}

		if err, ok := resp.ErrMap[recordID]; ok {
			h.Logger.WithFields(logrus.Fields{
				"recordID": recordID,
				"err":      err,
			}).Debugln("failed to delete record")
			result = serializedError{&recordID, err}
		} else {
			result = struct {
				RecordKey  string `json:"_recordID"`
				RecordType string `json:"_recordType"`
				Type       string `json:"_type"`
			}{recordID.Key, recordID.Type, "record"}
		}

		*results = append(*results, result)
	}
}
