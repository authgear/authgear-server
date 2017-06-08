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
		Convey("NewFieldACL should create a Field ACL with no entries", func() {
			acl := NewFieldACL(FieldACLEntryList{})
			So(acl, ShouldNotBeNil)
			So(len(acl.wildcardRecordType), ShouldEqual, 0)
			So(len(acl.recordTypes), ShouldEqual, 0)

		})

		Convey("NewFieldACL should create a Field ACL with entries", func() {
			entryList := []FieldACLEntry{
				{
					RecordType:   "*",
					RecordField:  "*",
					UserRole:     "_public",
					Writable:     true,
					Readable:     true,
					Comparable:   true,
					Discoverable: true,
				},
				{
					RecordType:   "note",
					RecordField:  "*",
					UserRole:     "_public",
					Writable:     true,
					Readable:     true,
					Comparable:   true,
					Discoverable: true,
				},
			}
			acl := NewFieldACL(FieldACLEntryList(entryList))
			So(acl, ShouldNotBeNil)
			So(acl.wildcardRecordType[0], ShouldResemble, entryList[0])
			So(acl.recordTypes["note"][0], ShouldResemble, entryList[1])
		})

		Convey("should check accessible for wildcard record type", func() {
			acl := NewFieldACL([]FieldACLEntry{
				{
					RecordType:   "*",
					RecordField:  "*",
					UserRole:     "_public",
					Writable:     true,
					Readable:     true,
					Comparable:   true,
					Discoverable: true,
				},
			})
			So(acl.Accessible(nil, nil, "note", "content", WriteFieldAccessMode), ShouldBeTrue)
		})

		Convey("should check accessible for specific record type", func() {
			acl := NewFieldACL(FieldACLEntryList{
				{
					RecordType:   "note",
					RecordField:  "*",
					UserRole:     "_public",
					Writable:     true,
					Readable:     true,
					Comparable:   true,
					Discoverable: true,
				},
			})
			So(acl.Accessible(nil, nil, "note", "content", WriteFieldAccessMode), ShouldBeTrue)
		})

		Convey("should returns false if entry list is empty", func() {
			acl := NewFieldACL(FieldACLEntryList{})
			So(acl.Accessible(nil, nil, "note", "content", WriteFieldAccessMode), ShouldBeFalse)
		})
	})
}

func TestFieldACLEntryList(t *testing.T) {
	Convey("FieldACLEntryList", t, func() {
		Convey("Accessible", func() {
			Convey("should return false for empty list", func() {
				entryList := FieldACLEntryList{}
				So(entryList.Accessible(nil, nil, "note", "content", WriteFieldAccessMode), ShouldBeFalse)
			})

			Convey("should return true if a entry is accessible", func() {
				entryList := FieldACLEntryList{
					{
						RecordType:   "note",
						RecordField:  "*",
						UserRole:     "_public",
						Writable:     true,
						Readable:     true,
						Comparable:   true,
						Discoverable: true,
					},
				}
				So(entryList.Accessible(nil, nil, "note", "content", WriteFieldAccessMode), ShouldBeTrue)
			})
		})

		Convey("Sort", func() {
			Convey("should sort", func() {
				entries := FieldACLEntryList{
					{"note", "*", "_public", true, true, false, false},
					{"*", "content", "_any_user", false, false, true, true},
					{"*", "*", "_public", false, false, false, false},
				}
				sort.Stable(entries)
				So(entries, ShouldResemble, FieldACLEntryList{
					{"*", "*", "_public", false, false, false, false},
					{"*", "content", "_any_user", false, false, true, true},
					{"note", "*", "_public", true, true, false, false},
				})
			})
		})
	})
}

func TestFieldACLEntry(t *testing.T) {
	Convey("FieldACLEntry", t, func() {
		entry := FieldACLEntry{
			RecordType:   "note",
			RecordField:  "*",
			UserRole:     "_public",
			Writable:     true,
			Readable:     true,
			Comparable:   true,
			Discoverable: true,
		}

		// TODO: add test case
		Convey("should return true", func() {
			So(entry.Accessible(nil, nil, "note", "content", WriteFieldAccessMode), ShouldBeTrue)
		})
	})
}
