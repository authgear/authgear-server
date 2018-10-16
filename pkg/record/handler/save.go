package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	recordGear "github.com/skygeario/skygear-server/pkg/record"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
	"github.com/skygeario/skygear-server/pkg/server/asset"
	"github.com/skygeario/skygear-server/pkg/server/plugin/hook"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/skyconv"
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

type RecordModifyRequest struct {
	RecordStore   record.Store
	AssetStore    asset.Store
	HookRegistry  *hook.Registry
	Atomic        bool
	WithMasterKey bool
	Context       context.Context
	AuthInfo      *skydb.AuthInfo
	ModifyAt      time.Time

	// Save only
	RecordsToSave []*skydb.Record

	// Delete Only
	RecordIDsToDelete []skydb.RecordID
}

type RecordModifyResponse struct {
	ErrMap           map[skydb.RecordID]skyerr.Error
	SavedRecords     []*skydb.Record
	DeletedRecordIDs []skydb.RecordID
}

type RecordFetcher struct {
	recordStore            record.Store
	withMasterKey          bool
	creationAccessCacheMap map[string]skydb.RecordACL
	defaultAccessCacheMap  map[string]skydb.RecordACL
	context                context.Context
	logger                 *logrus.Logger
}

// NewRecordFetcher provide a convenient FetchOrCreateRecord method
func NewRecordFetcher(ctx context.Context, recordStore record.Store, withMasterKey bool) RecordFetcher {
	return RecordFetcher{
		recordStore:            recordStore,
		withMasterKey:          withMasterKey,
		creationAccessCacheMap: map[string]skydb.RecordACL{},
		defaultAccessCacheMap:  map[string]skydb.RecordACL{},
		context:                ctx,
	}
}

func (f RecordFetcher) getCreationAccess(recordType string) skydb.RecordACL {
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

func (f RecordFetcher) getDefaultAccess(recordType string) skydb.RecordACL {
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

func (f RecordFetcher) FetchRecord(recordID skydb.RecordID, authInfo *skydb.AuthInfo, accessLevel skydb.RecordACLLevel) (record *skydb.Record, err skyerr.Error) {
	dbRecord := skydb.Record{}
	if dbErr := f.recordStore.Get(recordID, &dbRecord); dbErr != nil {
		if dbErr == skydb.ErrRecordNotFound {
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

	record = &dbRecord
	if !f.withMasterKey && !dbRecord.Accessible(authInfo, accessLevel) {
		err = skyerr.NewError(
			skyerr.PermissionDenied,
			"no permission to perform operation",
		)
	}

	return
}

func (f RecordFetcher) FetchOrCreateRecord(recordID skydb.RecordID, authInfo *skydb.AuthInfo) (record skydb.Record, created bool, err skyerr.Error) {
	fetchedRecord, err := f.FetchRecord(recordID, authInfo, skydb.WriteLevel)
	if err == nil {
		record = *fetchedRecord
		return
	}

	if err.Code() == skyerr.ResourceNotFound {
		allowCreation := func() bool {
			if f.withMasterKey {
				return true
			}

			creationAccess := f.getCreationAccess(recordID.Type)
			return creationAccess.Accessible(authInfo, skydb.CreateLevel)
		}()

		if !allowCreation {
			err = skyerr.NewError(
				skyerr.PermissionDenied,
				"no permission to create",
			)
			return
		}

		record = skydb.Record{}
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

	fetcher := NewRecordFetcher(req.Context, req.RecordStore, req.WithMasterKey)
	fieldACL, err := req.RecordStore.GetRecordFieldAccess()
	if err != nil {
		return skyerr.MakeError(err)
	}

	// fetch records
	originalRecordMap := map[skydb.RecordID]*skydb.Record{}
	records = executeRecordFunc(records, resp.ErrMap, func(record *skydb.Record) (err skyerr.Error) {
		dbRecord, created, err := fetcher.FetchOrCreateRecord(record.ID, req.AuthInfo)
		if err != nil {
			return err
		}

		now := req.ModifyAt
		if created {
			dbRecord.ID = record.ID
			dbRecord.OwnerID = req.AuthInfo.ID
			dbRecord.CreatedAt = now
			dbRecord.CreatorID = req.AuthInfo.ID
			dbRecord.UpdatedAt = now
			dbRecord.UpdaterID = req.AuthInfo.ID
		}

		if !req.WithMasterKey {
			if err = scrubRecordFieldsForWrite(
				req.AuthInfo,
				record,
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

		dbRecord.Apply(record)
		*record = dbRecord
		record.UpdatedAt = now
		record.UpdaterID = req.AuthInfo.ID

		return
	})

	// Apply default access
	records = executeRecordFunc(records, resp.ErrMap, func(record *skydb.Record) skyerr.Error {
		if record.ACL == nil {
			defaultACL := fetcher.getDefaultAccess(record.ID.Type)
			record.ACL = defaultACL
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
	records = executeRecordFunc(records, resp.ErrMap, func(record *skydb.Record) (err skyerr.Error) {
		var deltaRecord skydb.Record
		originalRecord, _ := originalRecordMap[record.ID]
		DeriveDeltaRecord(&deltaRecord, originalRecord, record)

		if dbErr := req.RecordStore.Save(&deltaRecord); dbErr != nil {
			err = skyerr.MakeError(dbErr)
		}
		*record = deltaRecord

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

type recordFunc func(*skydb.Record) skyerr.Error

func executeRecordFunc(recordsIn []*skydb.Record, errMap map[skydb.RecordID]skyerr.Error, rFunc recordFunc) (recordsOut []*skydb.Record) {
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
func scrubRecordFieldsForWrite(authInfo *skydb.AuthInfo, record *skydb.Record, origRecord *skydb.Record, fieldACL skydb.FieldACL, atomic bool) skyerr.Error {
	nonWritableFields := []string{}

	var deltaRecord skydb.Record
	DeriveDeltaRecord(&deltaRecord, origRecord, record)

	for key := range deltaRecord.Data {
		if fieldACL.Accessible(record.ID.Type, key, skydb.WriteFieldAccessMode, authInfo, origRecord) {
			continue
		}

		if atomic {
			nonWritableFields = append(nonWritableFields, key)
			continue
		}

		record.Remove(key)
	}

	if len(nonWritableFields) > 0 {
		return skyerr.NewDeniedArgument("Unable to save to some record fields because of Field ACL denied update.", nonWritableFields)
	}
	return nil
}

func injectSigner(record *skydb.Record, store asset.Store) {
	for _, value := range record.Data {
		switch v := value.(type) {
		case *skydb.Asset:
			if signer, ok := store.(asset.URLSigner); ok {
				v.Signer = signer
			} else {
				logrus.Warnf("Failed to acquire asset URLSigner, please check configuration")
			}
		}
	}
}

func makeAssetsCompleteAndInjectSigner(recordStore record.Store, records []*skydb.Record, store asset.Store) error {
	recordArr := []skydb.Record{}
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

func MakeAssetsComplete(recordStore record.Store, records []skydb.Record) error {
	if len(records) == 0 {
		return nil
	}

	recordType := records[0].ID.Type
	typemap, _ := recordStore.GetSchema(recordType)
	assetColumns := []string{}
	assetNames := []string{}

	for column, schema := range typemap {
		if schema.Type == skydb.TypeAsset {
			assetColumns = append(assetColumns, column)
		}
	}

	for _, record := range records {
		for _, assetColumn := range assetColumns {
			if thisAsset, ok := record.Get(assetColumn).(*skydb.Asset); ok {
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

	assetsByName := map[string]skydb.Asset{}
	for _, asset := range assets {
		assetsByName[asset.Name] = asset
	}
	for _, record := range records {
		for _, assetColumn := range assetColumns {
			if thisAsset, ok := record.Get(assetColumn).(*skydb.Asset); ok {
				completeAsset := assetsByName[thisAsset.Name]
				record.Set(assetColumn, &completeAsset)
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
func DeriveDeltaRecord(dst, base, delta *skydb.Record) {
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

func removeRecordFieldTypeHints(r *skydb.Record) {
	for k, v := range r.Data {
		switch v.(type) {
		case skydb.Sequence:
			delete(r.Data, k)
		case skydb.Unknown:
			delete(r.Data, k)
		}
	}
}
