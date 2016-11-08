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

package skydb

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRecord(t *testing.T) {
	Convey("Set transient field", t, func() {
		note0 := Record{
			ID: NewRecordID("note", "0"),
			Transient: Data{
				"content": "hello world",
			},
		}

		So(note0.Get("content"), ShouldBeNil)
		So(note0.Get("_transient"), ShouldResemble, Data{
			"content": "hello world",
		})
		So(note0.Get("_transient_content"), ShouldEqual, "hello world")
	})

	Convey("Set transient field", t, func() {
		note0 := Record{
			ID: NewRecordID("note", "0"),
		}

		note0.Set("_transient", Data{
			"content": "hello world",
		})

		So(note0.Data["content"], ShouldBeNil)
		So(note0.Transient, ShouldResemble, Data{
			"content": "hello world",
		})
	})

	Convey("Set individual transient field", t, func() {
		note0 := Record{
			ID: NewRecordID("note", "0"),
			Transient: Data{
				"existing": "should be here",
			},
		}

		note0.Set("_transient_content", "hello world")

		So(note0.Data["content"], ShouldBeNil)
		So(note0.Transient, ShouldResemble, Data{
			"content":  "hello world",
			"existing": "should be here",
		})
	})
}

func TestRecordACL(t *testing.T) {
	Convey("Record with ACL", t, func() {
		userinfo := &UserInfo{
			ID:    "user1",
			Roles: []string{"admin"},
		}

		stranger := &UserInfo{
			ID:    "stranger",
			Roles: []string{"nobody"},
		}
		Convey("Check public ace is pass on nil user", func() {
			ace := NewRecordACLEntryPublic(ReadLevel)

			So(ace.AccessibleLevel(ReadLevel), ShouldBeTrue)
			So(ace.Accessible(userinfo, ReadLevel), ShouldBeTrue)
			So(ace.Accessible(nil, ReadLevel), ShouldBeTrue)
		})

		Convey("Check public access right base on no user", func() {
			note := Record{
				ID:         NewRecordID("note", "0"),
				DatabaseID: "",
				ACL: RecordACL{
					NewRecordACLEntryPublic(ReadLevel),
				},
			}

			So(note.Accessible(userinfo, ReadLevel), ShouldBeTrue)
			So(note.Accessible(nil, ReadLevel), ShouldBeTrue)
		})

		Convey("Check access right base on role", func() {
			note := Record{
				ID:         NewRecordID("note", "0"),
				DatabaseID: "",
				ACL: RecordACL{
					NewRecordACLEntryRole("admin", ReadLevel),
				},
			}

			So(note.Accessible(userinfo, ReadLevel), ShouldBeTrue)
			So(note.Accessible(stranger, ReadLevel), ShouldBeFalse)
		})

		Convey("Check access right base on direct ace", func() {
			note := Record{
				ID:         NewRecordID("note", "0"),
				DatabaseID: "",
				ACL: RecordACL{
					NewRecordACLEntryDirect("user1", ReadLevel),
				},
			}

			So(note.Accessible(userinfo, ReadLevel), ShouldBeTrue)
			So(note.Accessible(stranger, ReadLevel), ShouldBeFalse)
		})

		Convey("Grant permission on any ACE matched", func() {
			note := Record{
				ID:         NewRecordID("note", "0"),
				DatabaseID: "",
				ACL: RecordACL{
					NewRecordACLEntryDirect("stranger", ReadLevel),
					NewRecordACLEntryRole("admin", ReadLevel),
				},
			}

			So(note.Accessible(userinfo, ReadLevel), ShouldBeTrue)
			So(note.Accessible(stranger, ReadLevel), ShouldBeTrue)
		})

		Convey("Write permission superset read permission", func() {
			note := Record{
				ID:         NewRecordID("note", "0"),
				DatabaseID: "",
				ACL: RecordACL{
					NewRecordACLEntryDirect("stranger", WriteLevel),
					NewRecordACLEntryRole("admin", WriteLevel),
				},
			}
			So(note.Accessible(userinfo, ReadLevel), ShouldBeTrue)
			So(note.Accessible(stranger, ReadLevel), ShouldBeTrue)
		})

		Convey("Reject write on read only permission", func() {
			note := Record{
				ID:         NewRecordID("note", "0"),
				DatabaseID: "",
				ACL: RecordACL{
					NewRecordACLEntryDirect("stranger", ReadLevel),
					NewRecordACLEntryRole("admin", ReadLevel),
				},
			}

			So(note.Accessible(userinfo, WriteLevel), ShouldBeFalse)
			So(note.Accessible(stranger, WriteLevel), ShouldBeFalse)
		})
	})
}

func TestRecordSchema(t *testing.T) {
	Convey("RecordSchema", t, func() {
		target := RecordSchema{
			"content": FieldType{TypeString, "", Expression{}},
			"date":    FieldType{TypeDateTime, "", Expression{}},
			"ref":     FieldType{TypeReference, "other", Expression{}},
		}

		Convey("is superset if equal", func() {
			other := RecordSchema{
				"content": FieldType{TypeString, "", Expression{}},
				"date":    FieldType{TypeDateTime, "", Expression{}},
				"ref":     FieldType{TypeReference, "other", Expression{}},
			}
			So(target.DefinitionSupersetOf(other), ShouldBeTrue)
		})

		Convey("is superset if target has all columns of the other schema", func() {
			other := RecordSchema{
				"content": FieldType{TypeString, "", Expression{}},
				"date":    FieldType{TypeDateTime, "", Expression{}},
			}
			So(target.DefinitionSupersetOf(other), ShouldBeTrue)
		})

		Convey("is not superset if wrong field type", func() {
			other := RecordSchema{
				"content": FieldType{TypeString, "", Expression{}},
				"date":    FieldType{TypeString, "", Expression{}},
				"ref":     FieldType{TypeReference, "other", Expression{}},
			}
			So(target.DefinitionSupersetOf(other), ShouldBeFalse)
		})

		Convey("is not superset if wrong reference type", func() {
			other := RecordSchema{
				"content": FieldType{TypeString, "", Expression{}},
				"date":    FieldType{TypeDateTime, "", Expression{}},
				"ref":     FieldType{TypeReference, "something", Expression{}},
			}
			So(target.DefinitionSupersetOf(other), ShouldBeFalse)
		})

		Convey("is not superset if column not exist in target", func() {
			other := RecordSchema{
				"content": FieldType{TypeString, "", Expression{}},
				"date":    FieldType{TypeDateTime, "", Expression{}},
				"ref":     FieldType{TypeReference, "other", Expression{}},
				"tag":     FieldType{TypeString, "", Expression{}},
			}
			So(target.DefinitionSupersetOf(other), ShouldBeFalse)
		})
	})
}
