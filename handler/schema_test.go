package handler

import (
	"encoding/json"
	"testing"

	"github.com/oursky/skygear/handler/handlertest"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skydb/skydbtest"
	. "github.com/oursky/skygear/skytest"
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

			result := &schemaResponse{}
			result.Encode(data)

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

			result := &schemaResponse{}
			result.Encode(data)

			expected := &schemaResponse{
				Schemas: map[string]schemaFieldList{
					"note": schemaFieldList{},
				},
			}
			So(result, ShouldResemble, expected)
		})
		Convey("empty record type", func() {
			data := map[string]skydb.RecordSchema{}

			result := &schemaResponse{}
			result.Encode(data)

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
			raw := []byte(`{"record_types":"something"}`)
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
		So(db.Extend("note", note), ShouldBeNil)

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
					"code":108,
					"message":"attempts to create reserved field",
					"name":"InvalidArgument"
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
				"error":{
					"code":114,
					"message":"Wrong type",
					"name":"IncompatibleSchema"
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
			raw := []byte(`{"record_types":"something"}`)
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
		So(db.Extend("note", note), ShouldBeNil)

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
					"code":108,
					"message":"attempts to change reserved key",
					"name":"InvalidArgument"
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
				"error":{
					"code":110,
					"message":"column notexist does not exist",
					"name":"ResourceNotFound"
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
				"error":{
					"code":110,
					"message":"column type conflict",
					"name":"ResourceNotFound"
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
			raw := []byte(`{"record_types":"something"}`)
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
		So(db.Extend("note", note), ShouldBeNil)

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
					"code":108,
					"message":"attempts to change reserved key",
					"name":"InvalidArgument"
				}
			}`)
		})

		Convey("delete nonexisting field", func() {
			resp := router.POST(`{
				"record_type": "note",
				"item_name": "notexist"
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error":{
					"code":110,
					"message":"column notexist does not exist",
					"name":"ResourceNotFound"
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
		So(db.Extend("note", note), ShouldBeNil)

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
