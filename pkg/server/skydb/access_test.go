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

		Convey("should return whether the access mode is accessible", func() {
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
			So(NewFieldUserRole("_public").Match(nil, nil), ShouldBeTrue)
			So(NewFieldUserRole("_any_user").Match(nil, nil), ShouldBeFalse)
			So(NewFieldUserRole("_user_id:janedoe").Match(&UserInfo{ID: "johndoe"}, nil), ShouldBeFalse)
			So(NewFieldUserRole("_user_id:johndoe").Match(&UserInfo{ID: "johndoe"}, nil), ShouldBeTrue)
			So(NewFieldUserRole("_role:admin").Match(&UserInfo{Roles: []string{"guest"}}, nil), ShouldBeFalse)
			So(NewFieldUserRole("_role:admin").Match(&UserInfo{Roles: []string{"admin"}}, nil), ShouldBeTrue)
			So(NewFieldUserRole("_any_user").Match(&UserInfo{}, nil), ShouldBeTrue)
		})
	})
}
