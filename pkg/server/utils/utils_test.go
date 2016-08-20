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

package utils

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestStringSliceExcept(t *testing.T) {
	Convey("StringSliceExcept", t, func() {
		Convey("return new slice without unwanted items", func() {
			result := StringSliceExcept([]string{
				"1",
				"2",
				"3",
			}, []string{
				"1",
				"3",
			})
			So(len(result), ShouldEqual, 1)
			So(result[0], ShouldEqual, "2")
		})

		Convey("should return all items if no items is filtered", func() {
			result := StringSliceExcept([]string{
				"1",
				"2",
				"3",
			}, []string{
				"4",
			})
			So(len(result), ShouldEqual, 3)
		})

		Convey("works with duplicated items to filter", func() {
			result := StringSliceExcept([]string{
				"1",
				"2",
				"3",
				"4",
				"5",
				"6",
				"7",
				"8",
				"9",
			}, []string{
				"4",
				"4",
				"1",
				"2",
				"3",
				"1",
				"2",
				"3",
				"7",
				"8",
				"9",
			})
			So(len(result), ShouldEqual, 2)
		})
	})
}

func TestStringSliceContainAny(t *testing.T) {
	Convey("StringSliceContainAny", t, func() {
		Convey("return true on container have any elements", func() {
			result := StringSliceContainAny([]string{
				"god",
				"man",
			}, []string{
				"god",
			})
			So(result, ShouldEqual, true)
		})
		Convey("return false on target slice is empty", func() {
			result := StringSliceContainAny([]string{
				"god",
				"man",
			}, []string{})
			So(result, ShouldEqual, false)
		})
		Convey("return false on container don't have all elements", func() {
			result := StringSliceContainAny([]string{
				"god",
				"man",
			}, []string{
				"devil",
			})
			So(result, ShouldEqual, false)
		})
	})
}
func TestStringSliceContainAll(t *testing.T) {
	Convey("StringSliceContainAll", t, func() {
		Convey("return true on container have all elements", func() {
			result := StringSliceContainAll([]string{
				"god",
				"man",
			}, []string{
				"god",
			})
			So(result, ShouldEqual, true)
		})
		Convey("return true on target slice is empty", func() {
			result := StringSliceContainAll([]string{
				"god",
				"man",
			}, []string{})
			So(result, ShouldEqual, true)
		})
		Convey("return false on container don't have all elements", func() {
			result := StringSliceContainAll([]string{
				"god",
				"man",
			}, []string{
				"devil",
			})
			So(result, ShouldEqual, false)
		})
	})
}
