package userprofile

import (
	"encoding/json"
	"time"

	"github.com/franela/goreq"
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/asset"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/model"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record" // tolerant nextimportslint: record
	recordHandler "github.com/skygeario/skygear-server/pkg/record/handler"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type recordStoreImpl struct {
	storeURL string
	apiKey   string
	logger   *logrus.Entry

	// followings are for record gear utilities
	authContext auth.ContextGetter
	txContext   db.TxContext
	recordStore record.Store
	assetStore  asset.Store
}

// NewUserProfileRecordStore returns a record-gear based user profile store implementation
func NewUserProfileRecordStore(
	storeURL string,
	apiKey string,
	logger *logrus.Entry,
	authContext auth.ContextGetter,
	txContext db.TxContext,
	recordStore record.Store,
	assetStore asset.Store,
) Store {
	return &recordStoreImpl{
		storeURL:    storeURL,
		apiKey:      apiKey,
		logger:      logger,
		authContext: authContext,
		txContext:   txContext,
		recordStore: recordStore,
	}
}

func (u *recordStoreImpl) CreateUserProfile(userID string, authInfo *authinfo.AuthInfo, data Data) (profile UserProfile, err error) {
	// for keeping integrity of authentication info (email, username, ...), writing on user record is banned by record gear,
	// so here, CreateUserProfile uses record gear utilities to save record directly rather than through public API.
	profileRecord := record.Record{
		ID: record.ID{
			Type: "user",
			Key:  userID,
		},
		OwnerID:   "",
		CreatedAt: time.Time{},
		CreatorID: "",
		UpdatedAt: time.Time{},
		UpdaterID: "",
		ACL:       nil,
		Data:      record.Data(data),
		Transient: nil,
	}
	recordsToSave := []*record.Record{
		&profileRecord,
	}
	modifyReq := recordHandler.RecordModifyRequest{
		RecordStore:   u.recordStore,
		TxContext:     u.txContext,
		AssetStore:    u.assetStore,
		Logger:        u.logger,
		AuthInfo:      authInfo,
		RecordsToSave: recordsToSave,
		Atomic:        true,
		WithMasterKey: u.authContext.AccessKeyType() == model.MasterAccessKey,
		ModifyAt:      timeNow(),
	}
	modifyResp := recordHandler.RecordModifyResponse{
		ErrMap: map[record.ID]skyerr.Error{},
	}

	// TODO: emit schema updated event
	_, err = recordHandler.ExtendRecordSchema(u.recordStore, u.logger, recordsToSave)
	if err != nil {
		u.logger.WithError(err).Errorln("failed to migrate profile record schema")
		if _, ok := err.(skyerr.Error); !ok {
			err = skyerr.NewError(skyerr.IncompatibleSchema, "failed to migrate profile record schema")
		}

		return
	}

	if err = recordHandler.RecordSaveHandler(&modifyReq, &modifyResp); err != nil {
		u.logger.WithError(err).Errorln("failed to save profile record")
		if _, ok := err.(skyerr.Error); !ok {
			err = skyerr.NewError(skyerr.IncompatibleSchema, "failed to migrate profile record schema")
		}

		return
	}

	respProfileRecord := modifyResp.SavedRecords[0]
	profile = UserProfile{
		Meta: Meta{
			ID:         respProfileRecord.ID.String(),
			Type:       "record",
			RecordID:   userID,
			RecordType: respProfileRecord.ID.Type,
			Access:     nil,
			OwnerID:    respProfileRecord.OwnerID,
			CreatedAt:  respProfileRecord.CreatedAt,
			CreatedBy:  respProfileRecord.CreatorID,
			UpdatedAt:  respProfileRecord.UpdatedAt,
			UpdatedBy:  respProfileRecord.UpdaterID,
		},
		Data: Data(respProfileRecord.Data),
	}
	return
}

func (u *recordStoreImpl) GetUserProfile(userID string, accessToken string) (profile UserProfile, err error) {
	body := make(map[string]interface{})
	body["record_type"] = "user"
	predicate := []interface{}{
		"eq",
		map[string]interface{}{
			"$val":  "_id",
			"$type": "keypath",
		},
		userID,
	}
	body["predicate"] = predicate

	resp, err := goreq.Request{
		Method: "POST",
		Uri:    u.storeURL + "query",
		Body:   body,
	}.
		WithHeader("X-Skygear-Api-Key", u.apiKey).
		WithHeader("X-Skygear-Access-Token", accessToken).
		Do()

	if err != nil {
		return
	}

	var bodyMap map[string]map[string][]Record
	err = resp.Body.FromJsonTo(&bodyMap)
	if err != nil {
		return
	}

	records, ok := bodyMap["result"]["records"]
	if !ok || len(records) < 1 {
		err = skyerr.NewError(skyerr.UnexpectedError, "Unable to fetch user profile")
		return
	}

	jsonRecord, err := json.Marshal(records[0])
	err = json.Unmarshal(jsonRecord, &profile)
	if err != nil {
		return
	}

	return
}
