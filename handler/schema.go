package handler

import (
	log "github.com/Sirupsen/logrus"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skyerr"
)

// SchemaRenameHandler handles the action of renaming column
type SchemaRenameHandler struct {
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectUser    router.Processor `preprocessor:"inject_user"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	RequireUser   router.Processor `preprocessor:"require_user"`
	preprocessors []router.Processor
}

func (h *SchemaRenameHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
		h.RequireUser,
	}
}

func (h *SchemaRenameHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *SchemaRenameHandler) Handle(payload *router.Payload, response *router.Response) {
	log.Debugf("%+v", payload)

	recordType, okType := payload.Data["record_type"].(string)
	oldName, okOldName := payload.Data["item_name"].(string)
	newName, okNewName := payload.Data["new_name"].(string)
	if !okType || !okOldName || !okNewName {
		response.Err = skyerr.NewError(skyerr.InvalidArgument, "data in the specified request is invalid")
		return
	}

	db := payload.Database

	if err := db.RenameSchema(recordType, oldName, newName); err != nil {
		response.Err = skyerr.NewError(skyerr.ResourceNotFound, err.Error())
		return
	}
	schema, err := db.FetchSchema(recordType)
	if err != nil {
		response.Err = skyerr.NewError(skyerr.UnexpectedError, err.Error())
		return
	}

	fields := []map[string]string{}
	for key, value := range schema {
		field := map[string]string{"name": key, "type": value.Type.String()}
		fields = append(fields, field)
	}
	results := map[string]interface{}{
		"record_types": map[string]interface{}{
			recordType: map[string]interface{}{
				"fields": fields,
			},
		},
	}

	response.Result = results
}

// SchemaDeleteHandler handles the action of deleting column
type SchemaDeleteHandler struct {
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectUser    router.Processor `preprocessor:"inject_user"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	RequireUser   router.Processor `preprocessor:"require_user"`
	preprocessors []router.Processor
}

func (h *SchemaDeleteHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
		h.RequireUser,
	}
}

func (h *SchemaDeleteHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *SchemaDeleteHandler) Handle(payload *router.Payload, response *router.Response) {
	log.Debugf("%+v", payload)

	recordType, okType := payload.Data["record_type"].(string)
	columnName, okColumnName := payload.Data["item_name"].(string)
	if !okType || !okColumnName {
		response.Err = skyerr.NewError(skyerr.InvalidArgument, "data in the specified request is invalid")
		return
	}

	db := payload.Database

	if err := db.DeleteSchema(recordType, columnName); err != nil {
		response.Err = skyerr.NewError(skyerr.ResourceNotFound, err.Error())
		return
	}
	schema, err := db.FetchSchema(recordType)
	if err != nil {
		response.Err = skyerr.NewError(skyerr.UnexpectedError, err.Error())
		return
	}

	fields := []map[string]string{}
	for key, value := range schema {
		field := map[string]string{"name": key, "type": value.Type.String()}
		fields = append(fields, field)
	}
	results := map[string]interface{}{
		"record_types": map[string]interface{}{
			recordType: map[string]interface{}{
				"fields": fields,
			},
		},
	}

	response.Result = results
}
