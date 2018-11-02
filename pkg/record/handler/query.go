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

type QueryResult struct {
	Records interface{}            `json:"records"`
	Info    map[string]interface{} `json:"info,omitempty"`
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
	parser := QueryParser{}
	authInfo := h.AuthContext.AuthInfo()
	if authInfo != nil {
		parser.UserID = authInfo.ID
	}
	data := map[string]interface{}{}
	json.NewDecoder(request.Body).Decode(&data)
	if err := parser.queryFromRaw(data, &payload.Query); err != nil {
		return nil, err
	}

	return payload, nil
}

func (h QueryHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(QueryRequestPayload)

	accessControlOptions := &record.AccessControlOptions{
		ViewAsUser:          h.AuthContext.AuthInfo(),
		BypassAccessControl: h.AuthContext.AccessKeyType() == model.MasterAccessKey,
	}

	fieldACL := func() record.FieldACL {
		acl, err := h.RecordStore.GetRecordFieldAccess()
		if err != nil {
			panic(err)
		}
		return acl
	}()

	if !accessControlOptions.BypassAccessControl {
		visitor := &queryAccessVisitor{
			FieldACL:   fieldACL,
			RecordType: payload.Query.Type,
			AuthInfo:   accessControlOptions.ViewAsUser,
			ExpressionACLChecker: ExpressionACLChecker{
				FieldACL:    fieldACL,
				RecordType:  payload.Query.Type,
				AuthInfo:    h.AuthContext.AuthInfo(),
				RecordStore: h.RecordStore,
			},
		}
		payload.Query.Accept(visitor)
		if err = visitor.Error(); err != nil {
			return
		}
	}

	results, err := h.RecordStore.Query(&payload.Query, accessControlOptions)
	if err != nil {
		err = skyerr.MakeError(err)
		return
	}
	defer results.Close()

	records := []record.Record{}
	for results.Scan() {
		record := results.Record()
		records = append(records, record)
	}

	err = results.Err()
	if err != nil {
		err = skyerr.MakeError(err)
		return
	}

	// Scan does not query assets,
	// it only replaces them with assets then only have name,
	// so we replace them with some complete assets.
	MakeAssetsComplete(h.RecordStore, records)

	eagerRecords := h.doQueryEager(h.eagerIDs(records, payload.Query), accessControlOptions)

	recordResultFilter, err := NewRecordResultFilter(
		h.RecordStore,
		h.TxContext,
		h.AssetStore,
		h.AuthContext.AuthInfo(),
		h.AuthContext.AccessKeyType() == model.MasterAccessKey,
	)
	if err != nil {
		err = skyerr.MakeError(err)
		return
	}

	result := QueryResult{}
	resultFilter := QueryResultFilter{
		RecordStore:        h.RecordStore,
		Query:              payload.Query,
		EagerRecords:       eagerRecords,
		RecordResultFilter: recordResultFilter,
	}

	output := make([]interface{}, len(records))
	for i := range records {
		record := records[i]
		output[i] = resultFilter.JSONResult(&record)
	}

	result.Records = output

	resultInfo, err := QueryResultInfo(h.RecordStore, &payload.Query, accessControlOptions, results)
	if err != nil {
		err = skyerr.MakeError(err)
		return
	}

	if len(resultInfo) > 0 {
		result.Info = resultInfo
	}

	resp = result

	return
}

func (h QueryHandler) eagerIDs(records []record.Record, query record.Query) map[string][]record.ID {
	eagers := map[string][]record.ID{}
	for _, transientExpression := range query.ComputedKeys {
		if transientExpression.Type != record.KeyPath {
			continue
		}
		keyPath := transientExpression.Value.(string)
		eagers[keyPath] = make([]record.ID, len(records))
	}

	for i, record := range records {
		for keyPath := range eagers {
			ref := getReferenceWithKeyPath(h.RecordStore, &record, keyPath)
			if ref.IsEmpty() {
				continue
			}
			eagers[keyPath][i] = ref.ID
		}
	}
	return eagers
}

func (h QueryHandler) doQueryEager(eagersIDs map[string][]record.ID, accessControlOptions *record.AccessControlOptions) map[string]map[string]*record.Record {
	eagerRecords := map[string]map[string]*record.Record{}
	for keyPath, ids := range eagersIDs {
		h.Logger.Debugf("Getting value for keypath %v", keyPath)
		eagerScanner, err := h.RecordStore.GetByIDs(ids, accessControlOptions)
		if err != nil {
			h.Logger.Debugf("No Records found in the eager load key path: %s", keyPath)
			eagerRecords[keyPath] = map[string]*record.Record{}
			continue
		}
		for eagerScanner.Scan() {
			er := eagerScanner.Record()
			if eagerRecords[keyPath] == nil {
				eagerRecords[keyPath] = map[string]*record.Record{}
			}
			eagerRecords[keyPath][er.ID.Key] = &er
		}
		eagerScanner.Close()
	}

	return eagerRecords
}
