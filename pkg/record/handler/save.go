package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/asset"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/model"
	"github.com/skygeario/skygear-server/pkg/core/server"
	recordGear "github.com/skygeario/skygear-server/pkg/record"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record/recordconv"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func AttachSaveHandler(
	server *server.Server,
	recordDependency recordGear.DependencyMap,
) *server.Server {
	server.Handle("/save", &RecordHandlerFactory{
		recordDependency,
	}).Methods("POST")
	return server
}

type RecordHandlerFactory struct {
	Dependency recordGear.DependencyMap
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
	Records []*record.Record

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
	AuthContext auth.ContextGetter `dependency:"AuthContextGetter"`
	TxContext   db.TxContext       `dependency:"TxContext"`
	RecordStore record.Store       `dependency:"RecordStore"`
	Logger      *logrus.Entry      `dependency:"HandlerLogger"`
	AssetStore  asset.Store        `dependency:"AssetStore"`
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
		var r record.Record
		if err := (*recordconv.JSONRecord)(&r).FromMap(recordMap); err != nil {
			skyErr := skyerr.NewError(skyerr.InvalidArgument, err.Error())
			payload.Errs = append(payload.Errs, skyErr)
			payload.IncomingItems = append(payload.IncomingItems, skyErr)
		} else {
			r.SanitizeForInput()
			payload.IncomingItems = append(payload.IncomingItems, r.ID)
			payload.Records = append(payload.Records, &r)
		}
	}

	return payload, nil
}

func (h SaveHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(SaveRequestPayload)

	resultFilter, err := NewRecordResultFilter(
		h.RecordStore,
		h.AssetStore,
		h.AuthContext.AuthInfo(),
		h.AuthContext.AccessKeyType() == model.MasterAccessKey,
	)
	if err != nil {
		err = skyerr.MakeError(err)
		return
	}

	modifyReq := RecordModifyRequest{
		RecordStore:   h.RecordStore,
		AssetStore:    h.AssetStore,
		Logger:        h.Logger,
		AuthInfo:      h.AuthContext.AuthInfo(),
		RecordsToSave: payload.Records,
		Atomic:        payload.Atomic,
		WithMasterKey: h.AuthContext.AccessKeyType() == model.MasterAccessKey,
		ModifyAt:      timeNow(),
	}
	modifyResp := RecordModifyResponse{
		ErrMap: map[record.ID]skyerr.Error{},
	}

	// TODO: emit schema updated event
	_, err = ExtendRecordSchema(h.RecordStore, h.Logger, payload.Records)
	if err != nil {
		h.Logger.WithError(err).Errorln("failed to migrate record schema")
		if myerr, ok := err.(skyerr.Error); ok {
			err = myerr
			return
		}

		err = skyerr.NewError(skyerr.IncompatibleSchema, "failed to migrate record schema")
		return
	}

	if err = RecordSaveHandler(&modifyReq, &modifyResp); err != nil {
		return
	}

	results := make([]interface{}, 0, len(payload.RawMaps))
	h.makeResultsFromIncomingItem(payload.IncomingItems, modifyResp, resultFilter, &results)

	resp = results

	return
}

func (h SaveHandler) makeResultsFromIncomingItem(incomingItems []interface{}, resp RecordModifyResponse, resultFilter RecordResultFilter, results *[]interface{}) {
	currRecordIdx := 0
	for _, itemi := range incomingItems {
		var result interface{}

		switch item := itemi.(type) {
		case skyerr.Error:
			result = newSerializedError("", item)
		case record.ID:
			if err, ok := resp.ErrMap[item]; ok {
				h.Logger.WithFields(logrus.Fields{
					"recordID": item,
					"err":      err,
				}).Debugln("failed to save record")

				result = newSerializedError(item.String(), err)
			} else {
				record := resp.SavedRecords[currRecordIdx]
				currRecordIdx++
				result = resultFilter.JSONResult(record)
			}
		default:
			panic(fmt.Sprintf("unknown type of incoming item: %T", itemi))
		}

		*results = append(*results, result)
	}
}

type RecordModifyRequest struct {
	RecordStore   record.Store
	AssetStore    asset.Store
	Logger        *logrus.Entry
	Atomic        bool
	WithMasterKey bool
	AuthInfo      *authinfo.AuthInfo
	ModifyAt      time.Time

	// Save only
	RecordsToSave []*record.Record

	// Delete Only
	RecordIDsToDelete []record.ID
}

type RecordModifyResponse struct {
	ErrMap           map[record.ID]skyerr.Error
	SavedRecords     []*record.Record
	DeletedRecordIDs []record.ID
}

type RecordFetcher struct {
	recordStore            record.Store
	withMasterKey          bool
	creationAccessCacheMap map[string]record.ACL
	defaultAccessCacheMap  map[string]record.ACL
	logger                 *logrus.Entry
}

// NewRecordFetcher provide a convenient FetchOrCreateRecord method
func NewRecordFetcher(recordStore record.Store, logger *logrus.Entry, withMasterKey bool) RecordFetcher {
	return RecordFetcher{
		recordStore:            recordStore,
		withMasterKey:          withMasterKey,
		creationAccessCacheMap: map[string]record.ACL{},
		defaultAccessCacheMap:  map[string]record.ACL{},
		logger:                 logger,
	}
}

func (f RecordFetcher) getCreationAccess(recordType string) record.ACL {
	creationAccess, creationAccessCached := f.creationAccessCacheMap[recordType]
	if creationAccessCached == false {
		var err error
		creationAccess, err = f.recordStore.GetRecordAccess(recordType)

		if err == nil && creationAccess != nil {
			f.creationAccessCacheMap[recordType] = creationAccess
		}
	}

	return creationAccess
}

func (f RecordFetcher) getDefaultAccess(recordType string) record.ACL {
	defaultAccess, defaultAccessCached := f.defaultAccessCacheMap[recordType]
	if defaultAccessCached == false {
		var err error
		defaultAccess, err = f.recordStore.GetRecordDefaultAccess(recordType)

		if err == nil && defaultAccess != nil {
			f.defaultAccessCacheMap[recordType] = defaultAccess
		}
	}

	return defaultAccess
}

func (f RecordFetcher) FetchRecord(recordID record.ID, authInfo *authinfo.AuthInfo, accessLevel record.ACLLevel) (r *record.Record, err skyerr.Error) {
	dbRecord := record.Record{}
	if dbErr := f.recordStore.Get(recordID, &dbRecord); dbErr != nil {
		if dbErr == record.ErrRecordNotFound {
			err = skyerr.NewError(skyerr.ResourceNotFound, "record not found")
		} else {
			f.logger.WithFields(logrus.Fields{
				"recordID": recordID,
				"err":      dbErr,
			}).Errorln("Failed to fetch record")
			err = skyerr.NewResourceFetchFailureErr("record", recordID.String())
		}
		return
	}

	r = &dbRecord
	if !f.withMasterKey && !dbRecord.Accessible(authInfo, accessLevel) {
		err = skyerr.NewError(
			skyerr.PermissionDenied,
			"no permission to perform operation",
		)
	}

	return
}

func (f RecordFetcher) FetchOrCreateRecord(recordID record.ID, authInfo *authinfo.AuthInfo) (r record.Record, created bool, err skyerr.Error) {
	fetchedRecord, err := f.FetchRecord(recordID, authInfo, record.WriteLevel)
	if err == nil {
		r = *fetchedRecord
		return
	}

	if err.Code() == skyerr.ResourceNotFound {
		allowCreation := func() bool {
			if f.withMasterKey {
				return true
			}

			creationAccess := f.getCreationAccess(recordID.Type)
			return creationAccess.Accessible(authInfo, record.CreateLevel)
		}()

		if !allowCreation {
			err = skyerr.NewError(
				skyerr.PermissionDenied,
				"no permission to create",
			)
			return
		}

		r = record.Record{}
		created = true
		err = nil
	}

	return
}

// RecordSaveHandler iterate the record to perform the following:
// 1. Query the db for original record
// 2. Execute before save hooks with original record and new record
// 3. Clean up some transport only data (sequence for example) away from record
// 4. Populate meta data and save the record (like updated_at/by)
// 5. Execute after save hooks with original record and new record
func RecordSaveHandler(req *RecordModifyRequest, resp *RecordModifyResponse) skyerr.Error {
	records := req.RecordsToSave

	fetcher := NewRecordFetcher(req.RecordStore, req.Logger, req.WithMasterKey)
	fieldACL, err := req.RecordStore.GetRecordFieldAccess()
	if err != nil {
		return skyerr.MakeError(err)
	}

	// fetch records
	originalRecordMap := map[record.ID]*record.Record{}
	records = executeRecordFunc(records, resp.ErrMap, func(r *record.Record) (err skyerr.Error) {
		dbRecord, created, err := fetcher.FetchOrCreateRecord(r.ID, req.AuthInfo)
		if err != nil {
			return err
		}

		now := req.ModifyAt
		if created {
			dbRecord.ID = r.ID
			dbRecord.OwnerID = req.AuthInfo.ID
			dbRecord.CreatedAt = now
			dbRecord.CreatorID = req.AuthInfo.ID
			dbRecord.UpdatedAt = now
			dbRecord.UpdaterID = req.AuthInfo.ID
		}

		if !req.WithMasterKey {
			if err = scrubRecordFieldsForWrite(
				req.AuthInfo,
				r,
				&dbRecord,
				fieldACL,
				req.Atomic,
			); err != nil {
				return
			}
		}

		if !created {
			origRecord := dbRecord.Copy()
			injectSigner(&origRecord, req.AssetStore)
			originalRecordMap[origRecord.ID] = &origRecord
		}

		dbRecord.Apply(r)
		*r = dbRecord
		r.UpdatedAt = now
		r.UpdaterID = req.AuthInfo.ID

		return
	})

	// Apply default access
	records = executeRecordFunc(records, resp.ErrMap, func(r *record.Record) skyerr.Error {
		if r.ACL == nil {
			defaultACL := fetcher.getDefaultAccess(r.ID.Type)
			r.ACL = defaultACL
		}
		return nil
	})

	makeAssetsCompleteAndInjectSigner(req.RecordStore, records, req.AssetStore)

	// TODO: before save hook

	// remove bogus field, they are only for schema change
	for _, r := range records {
		removeRecordFieldTypeHints(r)
	}

	// save records
	records = executeRecordFunc(records, resp.ErrMap, func(r *record.Record) (err skyerr.Error) {
		var deltaRecord record.Record
		originalRecord, _ := originalRecordMap[r.ID]
		DeriveDeltaRecord(&deltaRecord, originalRecord, r)

		if dbErr := req.RecordStore.Save(&deltaRecord); dbErr != nil {
			err = skyerr.MakeError(dbErr)
		}
		*r = deltaRecord

		return
	})

	if req.Atomic && len(resp.ErrMap) > 0 {
		return skyerr.NewError(skyerr.UnexpectedError, "atomic operation failed")
	}

	makeAssetsCompleteAndInjectSigner(req.RecordStore, records, req.AssetStore)

	// TODO: after save hook

	resp.SavedRecords = records

	return nil
}

type recordFunc func(*record.Record) skyerr.Error

func executeRecordFunc(recordsIn []*record.Record, errMap map[record.ID]skyerr.Error, rFunc recordFunc) (recordsOut []*record.Record) {
	for _, record := range recordsIn {
		if err := rFunc(record); err != nil {
			errMap[record.ID] = err
		} else {
			recordsOut = append(recordsOut, record)
		}
	}

	return
}

// scrubRecordFieldsForWrite checks the field ACL for write access.
// Depending on whether the request is an atomic one, this function
// will either remove the fields if the user is not allowed access if atomic
// is false, or will return an error.
func scrubRecordFieldsForWrite(authInfo *authinfo.AuthInfo, r *record.Record, origRecord *record.Record, fieldACL record.FieldACL, atomic bool) skyerr.Error {
	nonWritableFields := []string{}

	var deltaRecord record.Record
	DeriveDeltaRecord(&deltaRecord, origRecord, r)

	for key := range deltaRecord.Data {
		if fieldACL.Accessible(r.ID.Type, key, record.WriteFieldAccessMode, authInfo, origRecord) {
			continue
		}

		if atomic {
			nonWritableFields = append(nonWritableFields, key)
			continue
		}

		r.Remove(key)
	}

	if len(nonWritableFields) > 0 {
		return skyerr.NewDeniedArgument("Unable to save to some record fields because of Field ACL denied update.", nonWritableFields)
	}
	return nil
}

func injectSigner(r *record.Record, store asset.Store) {
	for _, value := range r.Data {
		switch v := value.(type) {
		case *record.Asset:
			if signer, ok := store.(asset.URLSigner); ok {
				v.Signer = signer
			} else {
				logrus.Warnf("Failed to acquire asset URLSigner, please check configuration")
			}
		}
	}
}

func makeAssetsCompleteAndInjectSigner(recordStore record.Store, records []*record.Record, store asset.Store) error {
	recordArr := []record.Record{}
	for _, v := range records {
		recordArr = append(recordArr, *v)
	}
	err := MakeAssetsComplete(recordStore, recordArr)
	if err != nil {
		return err
	}
	for _, record := range records {
		injectSigner(record, store)
	}
	return nil
}

func MakeAssetsComplete(recordStore record.Store, records []record.Record) error {
	if len(records) == 0 {
		return nil
	}

	recordType := records[0].ID.Type
	typemap, _ := recordStore.GetSchema(recordType)
	assetColumns := []string{}
	assetNames := []string{}

	for column, schema := range typemap {
		if schema.Type == record.TypeAsset {
			assetColumns = append(assetColumns, column)
		}
	}

	for _, r := range records {
		for _, assetColumn := range assetColumns {
			if thisAsset, ok := r.Get(assetColumn).(*record.Asset); ok {
				assetNames = append(assetNames, thisAsset.Name)
			}
		}
	}

	if len(assetNames) == 0 {
		return nil
	}

	assets, err := recordStore.GetAssets(assetNames)
	if err != nil {
		return err
	}

	assetsByName := map[string]record.Asset{}
	for _, asset := range assets {
		assetsByName[asset.Name] = asset
	}
	for _, r := range records {
		for _, assetColumn := range assetColumns {
			if thisAsset, ok := r.Get(assetColumn).(*record.Asset); ok {
				completeAsset := assetsByName[thisAsset.Name]
				r.Set(assetColumn, &completeAsset)
			}
		}
	}
	return nil
}

// DeriveDeltaRecord derive fields in delta which is either new or different from base, and
// write them in dst.
//
// It is the caller's reponsibility to ensure that base and delta identify
// the same record
func DeriveDeltaRecord(dst, base, delta *record.Record) {
	if base == nil {
		*dst = *delta
		return
	}

	dst.ID = delta.ID
	if delta.ACL != nil {
		dst.ACL = delta.ACL
	} else {
		dst.ACL = base.ACL
	}
	dst.OwnerID = delta.OwnerID
	dst.CreatedAt = delta.CreatedAt
	dst.CreatorID = delta.CreatorID
	dst.UpdatedAt = delta.UpdatedAt
	dst.UpdaterID = delta.UpdaterID

	dst.Data = map[string]interface{}{}
	for key, value := range delta.Data {
		if baseValue, ok := base.Data[key]; ok {
			// TODO(limouren): might want comparison that performs better
			if !reflect.DeepEqual(value, baseValue) {
				dst.Data[key] = value
			}
		} else {
			dst.Data[key] = value
		}
	}
}

func removeRecordFieldTypeHints(r *record.Record) {
	for k, v := range r.Data {
		switch v.(type) {
		case record.Sequence:
			delete(r.Data, k)
		case record.Unknown:
			delete(r.Data, k)
		}
	}
}

// RecordResultFilter is for processing Record into results.
//
// 1. Apply field-based acl, remove fields that are not accessible to the
//    provided authInfo
// 2. Inject asset
// 3. Return JSONRecord that is a copy of passed in Record that is ready to
//    be serialized
type RecordResultFilter struct {
	AssetStore          asset.Store
	FieldACL            record.FieldACL
	AuthInfo            *authinfo.AuthInfo
	BypassAccessControl bool
}

// NewRecordResultFilter return a RecordResultFilter.
func NewRecordResultFilter(recordStore record.Store, assetStore asset.Store, authInfo *authinfo.AuthInfo, bypassAccessControl bool) (RecordResultFilter, error) {
	var (
		acl record.FieldACL
		err error
	)

	if !bypassAccessControl {
		acl, err = recordStore.GetRecordFieldAccess()
		if err != nil {
			return RecordResultFilter{}, err
		}
	}

	return RecordResultFilter{
		AssetStore:          assetStore,
		AuthInfo:            authInfo,
		FieldACL:            acl,
		BypassAccessControl: bypassAccessControl,
	}, nil
}

func (f *RecordResultFilter) JSONResult(r *record.Record) *recordconv.JSONRecord {
	if r == nil {
		return nil
	}

	recordCopy := r.Copy()
	if !f.BypassAccessControl {
		scrubRecordFieldsForRead(f.AuthInfo, &recordCopy, f.FieldACL)
	}
	injectSigner(r, f.AssetStore)
	return (*recordconv.JSONRecord)(&recordCopy)
}

// scrubRecordFieldsForRead checks the field ACL to remove the fields
// from a record.Record that the user is not allowed to read.
func scrubRecordFieldsForRead(authInfo *authinfo.AuthInfo, r *record.Record, fieldACL record.FieldACL) {
	for _, key := range r.UserKeys() {
		if !fieldACL.Accessible(r.ID.Type, key, record.ReadFieldAccessMode, authInfo, r) {
			r.Remove(key)
		}
	}
}

func ExtendRecordSchema(recordStore record.Store, logger *logrus.Entry, records []*record.Record) (bool, error) {
	recordSchemaMergerMap := map[string]schemaMerger{}
	for _, record := range records {
		recordType := record.ID.Type
		merger, ok := recordSchemaMergerMap[recordType]
		if !ok {
			merger = newSchemaMerger()
		}

		merger.Extend(deriveRecordSchema(record.Data))

		// The map hold the value of Schema Merger. After we have
		// updated the Schema Merger, we have to copy the value
		// of Schema Merger back to the map.
		recordSchemaMergerMap[recordType] = merger
	}

	extended := false
	for recordType, merger := range recordSchemaMergerMap {
		schema, err := merger.Schema()
		if err != nil {
			return false, err
		}

		schemaExtended, err := recordStore.Extend(recordType, schema)
		if err != nil {
			return false, err
		}
		if schemaExtended {
			logger.
				WithField("type", recordType).
				WithField("schema", schema).
				Info("Schema Extended")
			extended = true
		}
	}

	return extended, nil
}

type schemaMerger struct {
	finalSchema record.Schema
	err         error
}

func newSchemaMerger() schemaMerger {
	return schemaMerger{finalSchema: record.Schema{}}
}

func (m *schemaMerger) Extend(schema record.Schema) {
	if m.err != nil {
		return
	}

	for key, dataType := range schema {
		if originalType, ok := m.finalSchema[key]; ok {
			if originalType != dataType {
				m.err = fmt.Errorf("type conflict on column = %s, %#v -> %#v", key, originalType, dataType)
				return
			}
		}

		m.finalSchema[key] = dataType
	}
}

func (m schemaMerger) Schema() (record.Schema, error) {
	return m.finalSchema, m.err
}

func deriveRecordSchema(m record.Data) record.Schema {
	schema := record.Schema{}
	for key, value := range m {
		if value == nil {
			continue
		}

		fieldType, err := record.DeriveFieldType(value)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"key":   key,
				"value": value,
			}).Panicf("unable to derive record schema: %s", err)
		}
		schema[key] = fieldType
	}

	return schema
}
