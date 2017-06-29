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
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTraverseColumnTypes(t *testing.T) {
	Convey("TraverseColumnTypes", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		db := NewMockDatabase(ctrl)
		db.EXPECT().RemoteColumnTypes(gomock.Eq("note")).
			Return(
				RecordSchema{
					"index":    FieldType{Type: TypeInteger},
					"category": FieldType{Type: TypeReference, ReferenceType: "category"},
				}, nil,
			).AnyTimes()
		db.EXPECT().RemoteColumnTypes(gomock.Eq("category")).
			Return(
				RecordSchema{
					"star": FieldType{Type: TypeBoolean},
				}, nil,
			).AnyTimes()
		db.EXPECT().RemoteColumnTypes(gomock.Eq("photo")).
			Return(
				RecordSchema{}, errors.New("no record type"),
			).AnyTimes()

		Convey("should find FieldType with simple keypath", func() {
			fields, err := TraverseColumnTypes(db, "note", "index")
			So(err, ShouldBeNil)
			So(fields, ShouldResemble, []FieldType{
				{Type: TypeInteger},
			})
		})

		Convey("should find FieldType with reference type", func() {
			fields, err := TraverseColumnTypes(db, "note", "category")
			So(err, ShouldBeNil)
			So(fields, ShouldResemble, []FieldType{
				{Type: TypeReference, ReferenceType: "category"},
			})
		})

		Convey("should traverse to FieldType with multi keypath", func() {
			fields, err := TraverseColumnTypes(db, "note", "category.star")
			So(err, ShouldBeNil)
			So(fields, ShouldResemble, []FieldType{
				{Type: TypeReference, ReferenceType: "category"},
				{Type: TypeBoolean},
			})
		})

		Convey("should return error if traversing a non-reference field", func() {
			_, err := TraverseColumnTypes(db, "note", "index.name")
			So(err, ShouldNotBeNil)
		})

		Convey("should return error for non-existence record type ", func() {
			_, err := TraverseColumnTypes(db, "photo", "url")
			So(err, ShouldNotBeNil)
		})

		Convey("should return error for non-existence key path", func() {
			_, err := TraverseColumnTypes(db, "note", "photo")
			So(err, ShouldNotBeNil)
		})
	})
}
