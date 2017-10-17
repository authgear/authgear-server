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
	"sort"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFieldACL(t *testing.T) {
	Convey("FieldACL", t, func() {
		publicRole := FieldUserRole{PublicFieldUserRoleType, ""}

		Convey("NewFieldACL should create a Field ACL with no entries", func() {
			acl := NewFieldACL(FieldACLEntryList{})
			So(acl, ShouldNotBeNil)
			So(len(acl.recordTypes), ShouldEqual, 0)
		})

		Convey("NewFieldACL should create a Field ACL with entries", func() {
			entry1 := FieldACLEntry{
				RecordType:   "*",
				RecordField:  "*",
				UserRole:     publicRole,
				Writable:     true,
				Readable:     true,
				Comparable:   true,
				Discoverable: true,
			}
			entry2 := FieldACLEntry{
				RecordType:   "note",
				RecordField:  "*",
				UserRole:     publicRole,
				Writable:     true,
				Readable:     true,
				Comparable:   true,
				Discoverable: true,
			}
			entryList := []FieldACLEntry{entry1, entry2}
			acl := NewFieldACL(FieldACLEntryList(entryList))
			So(acl, ShouldNotBeNil)
			So(acl.recordTypes[WildcardRecordType][0], ShouldResemble, entry1)
			So(acl.recordTypes["note"][0], ShouldResemble, entry2)
		})

		Convey("should return whether the access mode is accessible for public", func() {
			acl := NewFieldACL(FieldACLEntryList{
				{
					RecordType:   "*",
					RecordField:  "*",
					UserRole:     publicRole,
					Writable:     true,
					Readable:     false,
					Comparable:   true,
					Discoverable: true,
				},
				{
					RecordType:   "note",
					RecordField:  "content",
					UserRole:     publicRole,
					Writable:     false,
					Readable:     true,
					Comparable:   true,
					Discoverable: true,
				},
				{
					RecordType:   "note",
					RecordField:  "*",
					UserRole:     publicRole,
					Writable:     true,
					Readable:     true,
					Comparable:   false,
					Discoverable: true,
				},
			})

			So(acl.Accessible("note", "content", WriteFieldAccessMode, nil, nil), ShouldBeFalse)
			So(acl.Accessible("note", "content", ReadFieldAccessMode, nil, nil), ShouldBeTrue)
			So(acl.Accessible("note", "favorite", CompareFieldAccessMode, nil, nil), ShouldBeFalse)
			So(acl.Accessible("article", "content", ReadFieldAccessMode, nil, nil), ShouldBeFalse)
			So(acl.Accessible("article", "content", DiscoverOrCompareFieldAccessMode, nil, nil), ShouldBeTrue)
		})

		Convey("should return whether the access mode is accessible for user and record", func() {
			acl := NewFieldACL(FieldACLEntryList{
				{
					RecordType:   "*",
					RecordField:  "*",
					UserRole:     publicRole,
					Writable:     true,
					Readable:     true,
					Comparable:   true,
					Discoverable: true,
				},
				{
					RecordType:   "note",
					RecordField:  "*",
					UserRole:     publicRole,
					Writable:     false,
					Readable:     false,
					Comparable:   false,
					Discoverable: false,
				},
				{
					RecordType:   "note",
					RecordField:  "*",
					UserRole:     FieldUserRole{OwnerFieldUserRoleType, ""},
					Writable:     true,
					Readable:     true,
					Comparable:   true,
					Discoverable: true,
				},
				{
					RecordType:   "note",
					RecordField:  "*",
					UserRole:     FieldUserRole{DefinedRoleFieldUserRoleType, "admin"},
					Writable:     true,
					Readable:     false,
					Comparable:   true,
					Discoverable: true,
				},
			})

			johndoe := &AuthInfo{
				ID:    "johndoe",
				Roles: []string{"guest"},
			}
			janedoe := &AuthInfo{
				ID:    "janedoe",
				Roles: []string{"admin"},
			}
			record := &Record{
				OwnerID: "johndoe",
			}

			So(acl.Accessible("note", "content", ReadFieldAccessMode, nil, record), ShouldBeFalse)
			So(acl.Accessible("note", "content", ReadFieldAccessMode, johndoe, record), ShouldBeTrue)
			So(acl.Accessible("note", "content", WriteFieldAccessMode, janedoe, record), ShouldBeTrue)
			So(acl.Accessible("note", "content", CompareFieldAccessMode, johndoe, nil), ShouldBeFalse)
			So(acl.Accessible("note", "content", CompareFieldAccessMode, janedoe, nil), ShouldBeTrue)
		})

		Convey("should return whether the access mode is accessible for user and record with and without wildcard", func() {
			acl := NewFieldACL(FieldACLEntryList{
				{
					RecordType:   "*",
					RecordField:  "*",
					UserRole:     publicRole,
					Writable:     false,
					Readable:     true,
					Comparable:   true,
					Discoverable: true,
				},
				{
					RecordType:   "note",
					RecordField:  "*",
					UserRole:     FieldUserRole{OwnerFieldUserRoleType, ""},
					Writable:     true,
					Readable:     false,
					Comparable:   true,
					Discoverable: true,
				},
				{
					RecordType:   "note",
					RecordField:  "content",
					UserRole:     FieldUserRole{DefinedRoleFieldUserRoleType, "guest"},
					Writable:     false,
					Readable:     true,
					Comparable:   false,
					Discoverable: false,
				},
				{
					RecordType:   "note",
					RecordField:  "content",
					UserRole:     FieldUserRole{OwnerFieldUserRoleType, ""},
					Writable:     true,
					Readable:     true,
					Comparable:   true,
					Discoverable: true,
				},
				{
					RecordType:   "note",
					RecordField:  "stars",
					UserRole:     FieldUserRole{DefinedRoleFieldUserRoleType, "admin"},
					Writable:     true,
					Readable:     false,
					Comparable:   true,
					Discoverable: true,
				},
				{
					RecordType:   "note",
					RecordField:  "stars",
					UserRole:     FieldUserRole{DefinedRoleFieldUserRoleType, "guest"},
					Writable:     false,
					Readable:     true,
					Comparable:   true,
					Discoverable: true,
				},
			})

			johndoe := &AuthInfo{
				ID:    "johndoe",
				Roles: []string{"guest"},
			}
			janedoe := &AuthInfo{
				ID:    "janedoe",
				Roles: []string{"admin"},
			}
			record := &Record{
				OwnerID: "janedoe",
			}

			So(acl.Accessible("note", "content", ReadFieldAccessMode, nil, record), ShouldBeFalse)
			So(acl.Accessible("note", "content", WriteFieldAccessMode, nil, record), ShouldBeFalse)
			So(acl.Accessible("note", "content", ReadFieldAccessMode, johndoe, record), ShouldBeTrue)
			So(acl.Accessible("note", "content", WriteFieldAccessMode, johndoe, record), ShouldBeFalse)
			So(acl.Accessible("note", "content", ReadFieldAccessMode, janedoe, record), ShouldBeTrue)
			So(acl.Accessible("note", "content", WriteFieldAccessMode, janedoe, record), ShouldBeTrue)

			So(acl.Accessible("note", "title", ReadFieldAccessMode, nil, record), ShouldBeFalse)
			So(acl.Accessible("note", "title", WriteFieldAccessMode, nil, record), ShouldBeFalse)
			So(acl.Accessible("note", "title", ReadFieldAccessMode, johndoe, record), ShouldBeFalse)
			So(acl.Accessible("note", "title", WriteFieldAccessMode, johndoe, record), ShouldBeFalse)
			So(acl.Accessible("note", "title", ReadFieldAccessMode, janedoe, record), ShouldBeFalse)
			So(acl.Accessible("note", "title", WriteFieldAccessMode, janedoe, record), ShouldBeTrue)

			So(acl.Accessible("project", "title", ReadFieldAccessMode, nil, record), ShouldBeTrue)
			So(acl.Accessible("project", "title", WriteFieldAccessMode, nil, record), ShouldBeFalse)
			So(acl.Accessible("project", "title", ReadFieldAccessMode, johndoe, record), ShouldBeTrue)
			So(acl.Accessible("project", "title", WriteFieldAccessMode, johndoe, record), ShouldBeFalse)
			So(acl.Accessible("project", "title", ReadFieldAccessMode, janedoe, record), ShouldBeTrue)
			So(acl.Accessible("project", "title", WriteFieldAccessMode, janedoe, record), ShouldBeFalse)
		})

		Convey("should return false if no entry matches for user role", func() {
			acl := NewFieldACL(FieldACLEntryList{
				{
					RecordType:   "*",
					RecordField:  "*",
					UserRole:     FieldUserRole{OwnerFieldUserRoleType, ""},
					Writable:     true,
					Readable:     true,
					Comparable:   true,
					Discoverable: true,
				},
			})

			So(acl.Accessible("note", "content", ReadFieldAccessMode, nil, nil), ShouldBeFalse)
			So(acl.Accessible("note", "content", WriteFieldAccessMode, nil, nil), ShouldBeFalse)
		})

		Convey("should return whether the access mode is accessible for user with both wildcard", func() {
			acl := NewFieldACL(FieldACLEntryList{
				{
					RecordType:   "*",
					RecordField:  "*",
					UserRole:     FieldUserRole{OwnerFieldUserRoleType, ""},
					Writable:     true,
					Readable:     true,
					Comparable:   true,
					Discoverable: true,
				},
			})
			johndoe := &AuthInfo{
				ID: "johndoe",
			}
			janedoe := &AuthInfo{
				ID: "janedoe",
			}
			record := &Record{
				OwnerID: "janedoe",
			}

			So(acl.Accessible("note", "content", ReadFieldAccessMode, johndoe, record), ShouldBeFalse)
			So(acl.Accessible("note", "content", WriteFieldAccessMode, johndoe, record), ShouldBeFalse)

			So(acl.Accessible("note", "content", ReadFieldAccessMode, janedoe, record), ShouldBeTrue)
			So(acl.Accessible("note", "content", WriteFieldAccessMode, janedoe, record), ShouldBeTrue)
		})

		Convey("should returns true if no entry matches", func() {
			acl := NewFieldACL(FieldACLEntryList{})
			So(acl.Accessible("note", "content", WriteFieldAccessMode, nil, nil), ShouldBeTrue)
		})
	})
}

func TestFieldACLEntryList(t *testing.T) {
	Convey("FieldACLEntryList", t, func() {
		publicRole := FieldUserRole{PublicFieldUserRoleType, ""}
		anyUserRole := FieldUserRole{AnyUserFieldUserRoleType, ""}

		Convey("Sort", func() {
			Convey("should sort", func() {
				entries := FieldACLEntryList{
					{"note", "*", publicRole, true, true, false, false},
					{"*", "content", anyUserRole, false, false, true, true},
					{"*", "*", publicRole, false, false, false, false},
				}
				sort.Stable(entries)
				So(entries, ShouldResemble, FieldACLEntryList{
					{"note", "*", publicRole, true, true, false, false},
					{"*", "content", anyUserRole, false, false, true, true},
					{"*", "*", publicRole, false, false, false, false},
				})
			})
		})
	})
}

func TestFieldACLEntry(t *testing.T) {
	Convey("FieldACLEntry", t, func() {
		publicRole := FieldUserRole{PublicFieldUserRoleType, ""}

		entry := FieldACLEntry{
			RecordType:   "note",
			RecordField:  "*",
			UserRole:     publicRole,
			Writable:     true,
			Readable:     false,
			Comparable:   false,
			Discoverable: true,
		}

		Convey("should check accessible", func() {
			So(entry.Accessible(WriteFieldAccessMode), ShouldBeTrue)
			So(entry.Accessible(ReadFieldAccessMode), ShouldBeFalse)
			So(entry.Accessible(CompareFieldAccessMode), ShouldBeFalse)
			So(entry.Accessible(DiscoverOrCompareFieldAccessMode), ShouldBeTrue)
		})

		Convey("should compare entries", func() {
			compare := func(type1, field1, role1, type2, field2, role2 string) int {
				entry1 := FieldACLEntry{
					RecordType:  type1,
					RecordField: field1,
					UserRole:    NewFieldUserRole(role1),
				}
				entry2 := FieldACLEntry{
					RecordType:  type2,
					RecordField: field2,
					UserRole:    NewFieldUserRole(role2),
				}
				return entry1.Compare(entry2)
			}

			So(compare("note", "*", "_public", "note", "*", "_public"), ShouldEqual, 0)
			So(compare("*", "*", "_public", "note", "*", "_public"), ShouldBeGreaterThan, 0)
			So(compare("*", "content", "_public", "*", "*", "_public"), ShouldBeLessThan, 0)
			So(compare("*", "*", "_public", "*", "*", "_any_user"), ShouldBeGreaterThan, 0)
		})
	})
}

func TestFieldUserRole(t *testing.T) {
	Convey("FieldUserRole", t, func() {
		Convey("NewFieldUserRole", func() {
			Convey("should create field user role", func() {
				So(
					NewFieldUserRole("_owner"),
					ShouldResemble,
					FieldUserRole{OwnerFieldUserRoleType, ""},
				)
				So(
					NewFieldUserRole("_user_id:johndoe"),
					ShouldResemble,
					FieldUserRole{SpecificUserFieldUserRoleType, "johndoe"},
				)
				So(
					NewFieldUserRole("_field:stared"),
					ShouldResemble,
					FieldUserRole{DynamicUserFieldUserRoleType, "stared"},
				)
				So(
					NewFieldUserRole("_role:admin"),
					ShouldResemble,
					FieldUserRole{DefinedRoleFieldUserRoleType, "admin"},
				)
				So(
					NewFieldUserRole("_any_user"),
					ShouldResemble,
					FieldUserRole{AnyUserFieldUserRoleType, ""},
				)
				So(
					NewFieldUserRole("_public"),
					ShouldResemble,
					FieldUserRole{PublicFieldUserRoleType, ""},
				)
			})

			Convey("should panic if role string is not recognized", func() {
				So(func() { NewFieldUserRole("") }, ShouldPanic)
				So(func() { NewFieldUserRole("_invalid") }, ShouldPanic)
				So(func() { NewFieldUserRole("_owner:me") }, ShouldPanic)
				So(func() { NewFieldUserRole("_user_id") }, ShouldPanic)
				So(func() { NewFieldUserRole("_field") }, ShouldPanic)
				So(func() { NewFieldUserRole("_role") }, ShouldPanic)
				So(func() { NewFieldUserRole("_any_user:any") }, ShouldPanic)
				So(func() { NewFieldUserRole("_public:public") }, ShouldPanic)
			})
		})

		Convey("should generate string representation", func() {
			So(
				FieldUserRole{OwnerFieldUserRoleType, ""}.String(),
				ShouldEqual,
				"_owner",
			)
			So(
				FieldUserRole{SpecificUserFieldUserRoleType, "johndoe"}.String(),
				ShouldEqual,
				"_user_id:johndoe",
			)
			So(
				FieldUserRole{DynamicUserFieldUserRoleType, "stared"}.String(),
				ShouldEqual,
				"_field:stared",
			)
			So(
				FieldUserRole{DefinedRoleFieldUserRoleType, "admin"}.String(),
				ShouldEqual,
				"_role:admin",
			)
			So(
				FieldUserRole{AnyUserFieldUserRoleType, ""}.String(),
				ShouldEqual,
				"_any_user",
			)
			So(
				FieldUserRole{PublicFieldUserRoleType, ""}.String(),
				ShouldEqual,
				"_public",
			)
		})

		Convey("should compare user roles", func() {
			compare := func(role1, role2 string) int {
				return NewFieldUserRole(role1).Compare(NewFieldUserRole(role2))
			}

			So(compare("_any_user", "_any_user"), ShouldEqual, 0)
			So(compare("_owner", "_user_id:johndoe"), ShouldBeLessThan, 0)
			So(compare("_user_id:johndoe", "_user_id:janedoe"), ShouldBeGreaterThan, 0)
			So(compare("_user_id:john_doe", "_field:stared"), ShouldBeLessThan, 0)
			So(compare("_field:stared", "_role:admin"), ShouldBeLessThan, 0)
			So(compare("_role:admin", "_any_user"), ShouldBeLessThan, 0)
			So(compare("_any_user", "_public"), ShouldBeLessThan, 0)
		})

		Convey("should match user role", func() {
			johndoe := &AuthInfo{
				ID:    "johndoe",
				Roles: []string{"guest"},
			}
			janedoe := &AuthInfo{
				ID:    "janedoe",
				Roles: []string{"admin"},
			}
			record := &Record{
				OwnerID: "johndoe",
				Data: map[string]interface{}{
					"uid":  "johndoe",
					"uids": []interface{}{"janedoe"},
				},
			}

			So(NewFieldUserRole("_public").Match(nil, nil), ShouldBeTrue)
			So(NewFieldUserRole("_any_user").Match(nil, nil), ShouldBeFalse)
			So(NewFieldUserRole("_role:admin").Match(johndoe, nil), ShouldBeFalse)
			So(NewFieldUserRole("_role:admin").Match(janedoe, nil), ShouldBeTrue)
			So(NewFieldUserRole("_field:uid").Match(johndoe, record), ShouldBeTrue)
			So(NewFieldUserRole("_field:uid").Match(johndoe, nil), ShouldBeFalse)
			So(NewFieldUserRole("_field:uid").Match(janedoe, record), ShouldBeFalse)
			So(NewFieldUserRole("_field:uids").Match(johndoe, record), ShouldBeFalse)
			So(NewFieldUserRole("_field:uids").Match(janedoe, record), ShouldBeTrue)
			So(NewFieldUserRole("_user_id:janedoe").Match(johndoe, nil), ShouldBeFalse)
			So(NewFieldUserRole("_user_id:johndoe").Match(johndoe, nil), ShouldBeTrue)
			So(NewFieldUserRole("_owner").Match(johndoe, record), ShouldBeTrue)
			So(NewFieldUserRole("_owner").Match(johndoe, nil), ShouldBeFalse)
			So(NewFieldUserRole("_owner").Match(janedoe, record), ShouldBeFalse)
		})
	})
}
