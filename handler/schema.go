// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handler

import (
	"sort"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skyerr"
)

type schemaResponse struct {
	Schemas map[string]schemaFieldList `json:"record_types"`
}

type schemaFieldList struct {
	Fields []schemaField `mapstructure:"fields" json:"fields"`
}

func (s schemaFieldList) Len() int {
	return len(s.Fields)
}

func (s schemaFieldList) Swap(i, j int) {
	s.Fields[i], s.Fields[j] = s.Fields[j], s.Fields[i]
}

func (s schemaFieldList) Less(i, j int) bool {
	return strings.Compare(s.Fields[i].Name, s.Fields[j].Name) < 0
}

type schemaField struct {
	Name     string `mapstructure:"name" json:"name"`
	TypeName string `mapstructure:"type" json:"type"`
}

func (resp *schemaResponse) Encode(data map[string]skydb.RecordSchema) {
	resp.Schemas = make(map[string]schemaFieldList)
	for recordType, schema := range data {
		fieldList := schemaFieldList{}
		for fieldName, val := range schema {
			if strings.HasPrefix(fieldName, "_") {
				continue
			}

			fieldList.Fields = append(fieldList.Fields, schemaField{
				Name:     fieldName,
				TypeName: val.ToSimpleName(),
			})
		}
		sort.Sort(fieldList)
		resp.Schemas[recordType] = fieldList
	}
}

/*
SchemaRenameHandler handles the action of renaming column
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/ <<EOF
{
	"master_key": "MASTER_KEY",
	"action": "schema:rename",
	"record_type": "student",
	"item_type": "field",
	"item_name": "score",
	"new_name": "exam_score"
}
EOF
*/
type SchemaRenameHandler struct {
	DevOnly       router.Processor `preprocessor:"dev_only"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	preprocessors []router.Processor
}

func (h *SchemaRenameHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.DevOnly,
		h.DBConn,
		h.InjectDB,
	}
}

func (h *SchemaRenameHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

type schemaRenamePayload struct {
	RecordType string `mapstructure:"record_type"`
	OldName    string `mapstructure:"item_name"`
	NewName    string `mapstructure:"new_name"`
}

func (payload *schemaRenamePayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *schemaRenamePayload) Validate() skyerr.Error {
	missingArgs := []string{}
	if payload.RecordType == "" {
		missingArgs = append(missingArgs, "record_type")
	}
	if payload.OldName == "" {
		missingArgs = append(missingArgs, "item_name")
	}
	if payload.NewName == "" {
		missingArgs = append(missingArgs, "new_name")
	}
	if len(missingArgs) > 0 {
		return skyerr.NewInvalidArgument("missing required fields", missingArgs)
	}
	if strings.HasPrefix(payload.RecordType, "_") {
		return skyerr.NewInvalidArgument("attempts to change reserved table", []string{"record_type"})
	}
	if strings.HasPrefix(payload.OldName, "_") {
		return skyerr.NewInvalidArgument("attempts to change reserved key", []string{"item_name"})
	}
	if strings.HasPrefix(payload.NewName, "_") {
		return skyerr.NewInvalidArgument("attempts to change reserved key", []string{"new_name"})
	}
	return nil
}

func (h *SchemaRenameHandler) Handle(rpayload *router.Payload, response *router.Response) {
	payload := &schemaRenamePayload{}
	skyErr := payload.Decode(rpayload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	db := rpayload.Database

	if err := db.RenameSchema(payload.RecordType, payload.OldName, payload.NewName); err != nil {
		response.Err = skyerr.NewError(skyerr.ResourceNotFound, err.Error())
		return
	}

	results, err := db.GetRecordSchemas()
	if err != nil {
		response.Err = skyerr.NewError(skyerr.UnexpectedError, err.Error())
		return
	}

	resp := &schemaResponse{}
	resp.Encode(results)

	response.Result = resp
}

/*
SchemaDeleteHandler handles the action of deleting column
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/ <<EOF
{
	"master_key": "MASTER_KEY",
	"action": "schema:delete",
	"record_type": "student",
	"item_type": "field",
	"item_name": "score"
}
EOF
*/
type SchemaDeleteHandler struct {
	DevOnly       router.Processor `preprocessor:"dev_only"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	preprocessors []router.Processor
}

func (h *SchemaDeleteHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.DevOnly,
		h.DBConn,
		h.InjectDB,
	}
}

func (h *SchemaDeleteHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

type schemaDeletePayload struct {
	RecordType string `mapstructure:"record_type"`
	ColumnName string `mapstructure:"item_name"`
}

func (payload *schemaDeletePayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *schemaDeletePayload) Validate() skyerr.Error {
	missingArgs := []string{}
	if payload.RecordType == "" {
		missingArgs = append(missingArgs, "record_type")
	}
	if payload.ColumnName == "" {
		missingArgs = append(missingArgs, "item_name")
	}
	if len(missingArgs) > 0 {
		return skyerr.NewInvalidArgument("missing required fields", missingArgs)
	}
	if strings.HasPrefix(payload.RecordType, "_") {
		return skyerr.NewInvalidArgument("attempts to change reserved table", []string{"record_type"})
	}

	if strings.HasPrefix(payload.ColumnName, "_") {
		return skyerr.NewInvalidArgument("attempts to change reserved key", []string{"item_name"})
	}
	return nil
}

func (h *SchemaDeleteHandler) Handle(rpayload *router.Payload, response *router.Response) {
	payload := &schemaDeletePayload{}
	skyErr := payload.Decode(rpayload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	db := rpayload.Database

	if err := db.DeleteSchema(payload.RecordType, payload.ColumnName); err != nil {
		response.Err = skyerr.NewError(skyerr.ResourceNotFound, err.Error())
		return
	}

	results, err := db.GetRecordSchemas()
	if err != nil {
		response.Err = skyerr.NewError(skyerr.UnexpectedError, err.Error())
		return
	}

	resp := &schemaResponse{}
	resp.Encode(results)

	response.Result = resp
}

/*
SchemaCreateHandler handles the action of creating new columns
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/ <<EOF
{
	"master_key": "MASTER_KEY",
	"action": "schema:create",
	"record_types":{
		"student": {
			"fields":[
				{"name": "age", "type": "number"},
				{"name": "nickname" "type": "string"}
			]
		}
	}
}
EOF
*/
type SchemaCreateHandler struct {
	DevOnly       router.Processor `preprocessor:"dev_only"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	preprocessors []router.Processor
}

func (h *SchemaCreateHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.DevOnly,
		h.DBConn,
		h.InjectDB,
	}
}

func (h *SchemaCreateHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

type schemaCreatePayload struct {
	RawSchemas map[string]schemaFieldList `mapstructure:"record_types"`

	Schemas map[string]skydb.RecordSchema
}

func (payload *schemaCreatePayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}

	payload.Schemas = make(map[string]skydb.RecordSchema)
	for recordType, schema := range payload.RawSchemas {
		payload.Schemas[recordType] = make(skydb.RecordSchema)
		for _, field := range schema.Fields {
			var err error
			payload.Schemas[recordType][field.Name], err = skydb.SimpleNameToFieldType(field.TypeName)
			if err != nil {
				return skyerr.NewInvalidArgument("unexpected field type", []string{field.TypeName})
			}
		}
	}

	return payload.Validate()
}

func (payload *schemaCreatePayload) Validate() skyerr.Error {
	for recordType, schema := range payload.Schemas {
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

func (h *SchemaCreateHandler) Handle(rpayload *router.Payload, response *router.Response) {
	log.Debugf("%+v\n", rpayload)

	payload := &schemaCreatePayload{}
	skyErr := payload.Decode(rpayload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	db := rpayload.Database

	for recordType, recordSchema := range payload.Schemas {
		err := db.Extend(recordType, recordSchema)
		if err != nil {
			response.Err = skyerr.NewError(skyerr.IncompatibleSchema, err.Error())
			return
		}
	}

	results, err := db.GetRecordSchemas()
	if err != nil {
		response.Err = skyerr.NewError(skyerr.UnexpectedError, err.Error())
		return
	}

	resp := &schemaResponse{}
	resp.Encode(results)

	response.Result = resp
}

/*
SchemaFetchHandler handles the action of returing information of record schema
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/ <<EOF
{
	"master_key": "MASTER_KEY",
	"action": "schema:fetch"
}
EOF
*/
type SchemaFetchHandler struct {
	DevOnly       router.Processor `preprocessor:"dev_only"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	preprocessors []router.Processor
}

func (h *SchemaFetchHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.DevOnly,
		h.DBConn,
		h.InjectDB,
	}
}

func (h *SchemaFetchHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *SchemaFetchHandler) Handle(rpayload *router.Payload, response *router.Response) {
	db := rpayload.Database

	results, err := db.GetRecordSchemas()
	if err != nil {
		response.Err = skyerr.NewError(skyerr.UnexpectedError, err.Error())
		return
	}

	resp := &schemaResponse{}
	resp.Encode(results)

	response.Result = resp
}
