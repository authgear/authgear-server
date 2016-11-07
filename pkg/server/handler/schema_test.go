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
	"encoding/json"
	"testing"

	"github.com/skygeario/skygear-server/pkg/server/handler/handlertest"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/skydbtest"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSchemaResponse(t *testing.T) {
	Convey("schemaResponse", t, func() {
		Convey("normal schema", func() {
			data := map[string]skydb.RecordSchema{
				"note": skydb.RecordSchema{
					"field1": skydb.FieldType{
						Type: skydb.TypeString,
					},
					"field2": skydb.FieldType{
						Type: skydb.TypeDateTime,
					},
				},
			}

			result := &schemaResponse{
				Schemas: encodeRecordSchemas(data),
			}

			expected := &schemaResponse{
				Schemas: map[string]schemaFieldList{
					"note": schemaFieldList{
						Fields: []schemaField{
							schemaField{Name: "field1", TypeName: "string"},
							schemaField{Name: "field2", TypeName: "datetime"},
						},
					},
				},
			}
			So(result, ShouldResemble, expected)
		})
		Convey("empty schema", func() {
			data := map[string]skydb.RecordSchema{
				"note": skydb.RecordSchema{},
			}

			result := &schemaResponse{
				Schemas: encodeRecordSchemas(data),
			}

			expected := &schemaResponse{
				Schemas: map[string]schemaFieldList{
					"note": schemaFieldList{},
				},
			}
			So(result, ShouldResemble, expected)
		})
		Convey("empty record type", func() {
			data := map[string]skydb.RecordSchema{}

			result := &schemaResponse{
				Schemas: encodeRecordSchemas(data),
			}

			expected := &schemaResponse{
				Schemas: map[string]schemaFieldList{},
			}
			So(result, ShouldResemble, expected)
		})
	})
}

func TestSchemaCreatePayload(t *testing.T) {
	Convey("SchemaCreatePayload", t, func() {
		payload := &schemaCreatePayload{}

		Convey("normal payload", func() {
			raw := []byte(`{
				"record_types": {
					"note": {
						"fields": [
							{"name": "field1", "type": "string"},
							{"name": "field2", "type": "datetime"}
						]
					}
				}
			}`)
			var data map[string]interface{}
			err := json.Unmarshal(raw, &data)
			So(err, ShouldBeNil)

			skyErr := payload.Decode(data)
			So(skyErr, ShouldBeNil)

			expected := &schemaCreatePayload{
				RawSchemas: map[string]schemaFieldList{
					"note": schemaFieldList{
						Fields: []schemaField{
							schemaField{Name: "field1", TypeName: "string"},
							schemaField{Name: "field2", TypeName: "datetime"},
						},
					},
				},

				Schemas: map[string]skydb.RecordSchema{
					"note": skydb.RecordSchema{
						"field1": skydb.FieldType{
							Type: skydb.TypeString,
						},
						"field2": skydb.FieldType{
							Type: skydb.TypeDateTime,
						},
					},
				},
			}

			So(payload, ShouldResemble, expected)
		})

		Convey("reserved field", func() {
			raw := []byte(`{
				"record_types": {
					"note": {
						"fields": [
							{"name": "_field1", "type": "string"},
							{"name": "field2", "type": "datetime"}
						]
					}
				}
			}`)
			var data map[string]interface{}
			err := json.Unmarshal(raw, &data)
			So(err, ShouldBeNil)

			skyErr := payload.Decode(data)
			So(skyErr, ShouldNotBeNil)
		})

		Convey("reserved type", func() {
			raw := []byte(`{
				"record_types": {
					"_note": {
						"fields": [
							{"name": "field1", "type": "string"},
							{"name": "field2", "type": "datetime"}
						]
					}
				}
			}`)
			var data map[string]interface{}
			err := json.Unmarshal(raw, &data)
			So(err, ShouldBeNil)

			skyErr := payload.Decode(data)
			So(skyErr, ShouldNotBeNil)
		})

		Convey("unsupported type", func() {
			raw := []byte(`{
				"record_types": {
					"note": {
						"fields": [
							{"name": "field1", "type": "unsupported"},
							{"name": "field2", "type": "datetime"}
						]
					}
				}
			}`)
			var data map[string]interface{}
			err := json.Unmarshal(raw, &data)
			So(err, ShouldBeNil)

			skyErr := payload.Decode(data)
			So(skyErr, ShouldNotBeNil)

		})

		Convey("wrong json", func() {
			raw := []byte(`{"record_types": "something"}`)
			var data map[string]interface{}
			err := json.Unmarshal(raw, &data)
			So(err, ShouldBeNil)

			skyErr := payload.Decode(data)
			So(skyErr, ShouldNotBeNil)
		})

	})
}

func TestSchemaCreateHandler(t *testing.T) {
	Convey("SchemaCreateHandler", t, func() {
		note := skydb.RecordSchema{
			"field1": skydb.FieldType{
				Type: skydb.TypeString,
			},
			"field2": skydb.FieldType{
				Type: skydb.TypeDateTime,
			},
		}

		db := skydbtest.NewMapDB()
		_, err := db.Extend("note", note)
		So(err, ShouldBeNil)

		router := handlertest.NewSingleRouteRouter(&SchemaCreateHandler{}, func(p *router.Payload) {
			p.Database = db
		})

		Convey("create normal field", func() {
			resp := router.POST(`{
				"record_types": {
					"note": {
						"fields": [
							{"name": "field3", "type": "string"},
							{"name": "field4", "type": "number"}
						]
					}
				}
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"record_types": {
						"note": {
							"fields": [
								{"name": "field1", "type": "string"},
								{"name": "field2", "type": "datetime"},
								{"name": "field3", "type": "string"},
								{"name": "field4", "type": "number"}
							]
						}
					}
				}
			}`)
		})

		Convey("create reserved field", func() {
			resp := router.POST(`{
				"record_types": {
					"note": {
						"fields": [
							{"name": "_field3", "type": "string"}
						]
					}
				}
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 108,
					"message": "attempts to create reserved field",
					"info": {
						"arguments": [
							"_field3"
						]
					},
					"name": "InvalidArgument"
				}
			}`)
		})

		Convey("create existing field with conflict", func() {
			resp := router.POST(`{
				"record_types": {
					"note": {
						"fields": [
							{"name": "field1", "type": "integer"}
						]
					}
				}
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 114,
					"message": "Wrong type",
					"name": "IncompatibleSchema"
				}
			}`)
		})

		Convey("create existing field without conflict", func() {
			resp := router.POST(`{
				"record_types": {
					"note": {
						"fields": [
							{"name": "field1", "type": "string"}
						]
					}
				}
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"record_types": {
						"note": {
							"fields": [
								{"name": "field1", "type": "string"},
								{"name": "field2", "type": "datetime"}
							]
						}
					}
				}
			}`)
		})

	})
}

func TestSchemaRenamePayload(t *testing.T) {
	Convey("SchemaRenamePayload", t, func() {
		payload := &schemaRenamePayload{}

		Convey("normal payload", func() {
			raw := []byte(`{
				"record_type": "note",
				"item_name": "field1",
				"new_name": "newName"
			}`)
			var data map[string]interface{}
			err := json.Unmarshal(raw, &data)
			So(err, ShouldBeNil)

			skyErr := payload.Decode(data)
			So(skyErr, ShouldBeNil)

			expected := &schemaRenamePayload{
				RecordType: "note",
				OldName:    "field1",
				NewName:    "newName",
			}

			So(payload, ShouldResemble, expected)
		})

		Convey("reserved field", func() {
			raw := []byte(`{
				"record_type": "note",
				"item_name": "_field1",
				"new_name": "newName"
			}`)
			var data map[string]interface{}
			err := json.Unmarshal(raw, &data)
			So(err, ShouldBeNil)

			skyErr := payload.Decode(data)
			So(skyErr, ShouldNotBeNil)
		})

		Convey("reserved new field", func() {
			raw := []byte(`{
				"record_type": "note",
				"item_name": "field1",
				"new_name": "_newName"
			}`)
			var data map[string]interface{}
			err := json.Unmarshal(raw, &data)
			So(err, ShouldBeNil)

			skyErr := payload.Decode(data)
			So(skyErr, ShouldNotBeNil)
		})

		Convey("reserved type", func() {
			raw := []byte(`{
				"record_type": "_note",
				"item_name": "field1",
				"new_name": "newName"
			}`)
			var data map[string]interface{}
			err := json.Unmarshal(raw, &data)
			So(err, ShouldBeNil)

			skyErr := payload.Decode(data)
			So(skyErr, ShouldNotBeNil)
		})

		Convey("wrong json", func() {
			raw := []byte(`{"record_types": "something"}`)
			var data map[string]interface{}
			err := json.Unmarshal(raw, &data)
			So(err, ShouldBeNil)

			skyErr := payload.Decode(data)
			So(skyErr, ShouldNotBeNil)
		})

	})
}

func TestSchemaRenameHandler(t *testing.T) {
	Convey("SchemaRenameHandler", t, func() {
		note := skydb.RecordSchema{
			"field1": skydb.FieldType{
				Type: skydb.TypeString,
			},
			"field2": skydb.FieldType{
				Type: skydb.TypeDateTime,
			},
		}

		db := skydbtest.NewMapDB()
		_, err := db.Extend("note", note)
		So(err, ShouldBeNil)

		router := handlertest.NewSingleRouteRouter(&SchemaRenameHandler{}, func(p *router.Payload) {
			p.Database = db
		})

		Convey("rename normal field", func() {
			resp := router.POST(`{
				"record_type": "note",
				"item_name": "field1",
				"new_name": "newName"
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"record_types": {
						"note": {
							"fields": [
								{"name": "field2", "type": "datetime"},
								{"name": "newName", "type": "string"}
							]
						}
					}
				}
			}`)
		})

		Convey("rename reserved field", func() {
			resp := router.POST(`{
				"record_type": "note",
				"item_name": "_id",
				"new_name": "newName"
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 108,
					"message": "attempts to change reserved key",
					"info": {
						"arguments": [
							"item_name"
						]
					},
					"name": "InvalidArgument"
				}
			}`)
		})

		Convey("rename nonexisting field", func() {
			resp := router.POST(`{
				"record_type": "note",
				"item_name": "notexist",
				"new_name": "newName"
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 110,
					"message": "column notexist does not exist",
					"name": "ResourceNotFound"
				}
			}`)
		})

		Convey("rename to existing field", func() {
			resp := router.POST(`{
				"record_type": "note",
				"item_name": "field1",
				"new_name": "field2"
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 110,
					"message": "column type conflict",
					"name": "ResourceNotFound"
				}
			}`)
		})
	})
}

func TestSchemaDeletePayload(t *testing.T) {
	Convey("SchemaDeletePayload", t, func() {
		payload := &schemaDeletePayload{}

		Convey("normal payload", func() {
			raw := []byte(`{
				"record_type": "note",
				"item_name": "field1"
			}`)
			var data map[string]interface{}
			err := json.Unmarshal(raw, &data)
			So(err, ShouldBeNil)

			skyErr := payload.Decode(data)
			So(skyErr, ShouldBeNil)

			expected := &schemaDeletePayload{
				RecordType: "note",
				ColumnName: "field1",
			}

			So(payload, ShouldResemble, expected)
		})

		Convey("reserved field", func() {
			raw := []byte(`{
				"record_type": "note",
				"item_name": "_field1"
			}`)
			var data map[string]interface{}
			err := json.Unmarshal(raw, &data)
			So(err, ShouldBeNil)

			skyErr := payload.Decode(data)
			So(skyErr, ShouldNotBeNil)
		})

		Convey("reserved type", func() {
			raw := []byte(`{
				"record_type": "_note",
				"item_name": "field1"
			}`)
			var data map[string]interface{}
			err := json.Unmarshal(raw, &data)
			So(err, ShouldBeNil)

			skyErr := payload.Decode(data)
			So(skyErr, ShouldNotBeNil)
		})

		Convey("wrong json", func() {
			raw := []byte(`{"record_types": "something"}`)
			var data map[string]interface{}
			err := json.Unmarshal(raw, &data)
			So(err, ShouldBeNil)

			skyErr := payload.Decode(data)
			So(skyErr, ShouldNotBeNil)
		})

	})
}

func TestSchemaDeleteHandler(t *testing.T) {
	Convey("SchemaDeleteHandler", t, func() {
		note := skydb.RecordSchema{
			"field1": skydb.FieldType{
				Type: skydb.TypeString,
			},
			"field2": skydb.FieldType{
				Type: skydb.TypeDateTime,
			},
		}

		db := skydbtest.NewMapDB()
		_, err := db.Extend("note", note)
		So(err, ShouldBeNil)

		router := handlertest.NewSingleRouteRouter(&SchemaDeleteHandler{}, func(p *router.Payload) {
			p.Database = db
		})

		Convey("delete normal field", func() {
			resp := router.POST(`{
				"record_type": "note",
				"item_name": "field1"
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"record_types": {
						"note": {
							"fields": [
								{"name": "field2", "type": "datetime"}
							]
						}
					}
				}
			}`)
		})

		Convey("delete reserved field", func() {
			resp := router.POST(`{
				"record_type": "note",
				"item_name": "_id"
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 108,
					"message": "attempts to change reserved key",
					"info": {
						"arguments": [
							"item_name"
						]
					},
					"name": "InvalidArgument"
				}
			}`)
		})

		Convey("delete nonexisting field", func() {
			resp := router.POST(`{
				"record_type": "note",
				"item_name": "notexist"
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 110,
					"message": "column notexist does not exist",
					"name": "ResourceNotFound"
				}
			}`)
		})
	})
}

func TestSchemaFetchHandler(t *testing.T) {
	Convey("SchemaFetchHandler", t, func() {
		note := skydb.RecordSchema{
			"field1": skydb.FieldType{
				Type: skydb.TypeString,
			},
			"field2": skydb.FieldType{
				Type: skydb.TypeDateTime,
			},
		}

		db := skydbtest.NewMapDB()
		_, err := db.Extend("note", note)
		So(err, ShouldBeNil)

		router := handlertest.NewSingleRouteRouter(&SchemaFetchHandler{}, func(p *router.Payload) {
			p.Database = db
		})

		Convey("fetch schemas", func() {
			resp := router.POST(`{}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"record_types": {
						"note": {
							"fields": [
								{"name": "field1", "type": "string"},
								{"name": "field2", "type": "datetime"}
							]
						}
					}
				}
			}`)
		})
	})
}

func TestSchemaAccessPayload(t *testing.T) {
	Convey("SchemaAccessPayload", t, func() {
		Convey("Valid Data", func() {
			payload := schemaAccessPayload{}
			skyErr := payload.Decode(map[string]interface{}{
				"action": "schema:access",
				"type":   "script",
				"create_roles": []string{
					"Admin",
					"Writer",
				},
			})

			So(skyErr, ShouldBeNil)
			So(payload.Validate(), ShouldBeNil)

			roleNames := []string{}
			for _, perACE := range payload.ACL {
				if perACE.Role != "" {
					roleNames = append(roleNames, perACE.Role)
				}
			}

			So(roleNames, ShouldContain, "Admin")
			So(roleNames, ShouldContain, "Writer")
		})

		Convey("Invalid Data", func() {
			payload := schemaAccessPayload{}
			err := payload.Decode(map[string]interface{}{
				"action": "schema:access",
				"create_roles": []string{
					"Admin",
					"Writer",
				},
			})

			So(err, ShouldResemble,
				skyerr.NewInvalidArgument("missing required fields", []string{"type"}))

			err = payload.Decode(map[string]interface{}{
				"action":       "schema:access",
				"type":         "script",
				"create_roles": "Admin",
			})

			So(err, ShouldResemble,
				skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload"))
		})
	})
}

type mockSchemaAccessDatabase struct {
	DBConn skydb.Conn

	skydb.Database
}

func (db *mockSchemaAccessDatabase) Conn() skydb.Conn {
	return db.DBConn
}

type mockSchemaAccessDatabaseConnection struct {
	recordType string
	acl        skydb.RecordACL

	skydb.Conn
}

func (c *mockSchemaAccessDatabaseConnection) SetRecordAccess(recordType string, acl skydb.RecordACL) error {
	c.recordType = recordType
	c.acl = acl

	return nil
}

func TestSchemaAccessHandler(t *testing.T) {
	Convey("TestSchemaAccessHandler", t, func() {
		mockConn := &mockSchemaAccessDatabaseConnection{}
		mockDB := &mockSchemaAccessDatabase{}
		mockDB.DBConn = mockConn

		handler := handlertest.NewSingleRouteRouter(&SchemaAccessHandler{}, func(p *router.Payload) {
			p.Database = mockDB
		})

		resp := handler.POST(`{
			"type": "script",
			"create_roles": ["Admin", "Writer"]
		}`)

		So(resp.Body.Bytes(), ShouldEqualJSON, `{
			"result": {
				"type": "script",
				"create_roles": ["Admin", "Writer"]
			}
		}`)

		So(mockConn.recordType, ShouldEqual, "script")

		roleNames := []string{}
		for _, perACE := range mockConn.acl {
			if perACE.Role != "" {
				roleNames = append(roleNames, perACE.Role)
			}
		}

		So(roleNames, ShouldContain, "Admin")
		So(roleNames, ShouldContain, "Writer")
	})
}
