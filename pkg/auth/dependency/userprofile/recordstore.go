package userprofile

import (
	"time"

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
			err = skyerr.NewError(skyerr.IncompatibleSchema, "failed to save profile record")
		}

		return
	}

	profile = u.toUserProfile(*modifyResp.SavedRecords[0])
	return
}

func (u *recordStoreImpl) GetUserProfile(userID string, accessToken string) (profile UserProfile, err error) {
	accessControlOptions := &record.AccessControlOptions{
		ViewAsUser:          u.authContext.AuthInfo(),
		BypassAccessControl: u.authContext.AccessKeyType() == model.MasterAccessKey,
	}
	predicate := record.Predicate{
		Operator: record.Equal,
		Children: make([]interface{}, 0),
	}
	predicate.Children = append(predicate.Children, record.Expression{
		Type:  record.KeyPath,
		Value: "_id",
	})
	predicate.Children = append(predicate.Children, record.Expression{
		Type:  record.Literal,
		Value: userID,
	})
	query := record.Query{
		Type:      "user",
		Predicate: predicate,
		Limit:     new(uint64),
	}
	*query.Limit = 1

	// TODO: maybe need ACL checking before query

	results, err := u.recordStore.Query(&query, accessControlOptions)
	if err != nil {
		u.logger.WithError(err).Errorln("failed to get profile record")
		if _, ok := err.(skyerr.Error); !ok {
			err = skyerr.NewError(skyerr.IncompatibleSchema, "failed to get profile record")
		}
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
		u.logger.WithError(err).Errorln("failed to scan profile record")
		if _, ok := err.(skyerr.Error); !ok {
			err = skyerr.NewError(skyerr.IncompatibleSchema, "failed to scan profile record")
		}
		return
	}

	profile = u.toUserProfile(records[0])
	return
}

func (u *recordStoreImpl) toUserProfile(profileReocrd record.Record) UserProfile {
	return UserProfile{
		Meta: Meta{
			ID:         profileReocrd.ID.String(),
			Type:       "record",
			RecordID:   profileReocrd.ID.Key,
			RecordType: profileReocrd.ID.Type,
			Access:     nil,
			OwnerID:    profileReocrd.OwnerID,
			CreatedAt:  profileReocrd.CreatedAt,
			CreatedBy:  profileReocrd.CreatorID,
			UpdatedAt:  profileReocrd.UpdatedAt,
			UpdatedBy:  profileReocrd.UpdaterID,
		},
		Data: Data(profileReocrd.Data),
	}
}
