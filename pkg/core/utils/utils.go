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

// StringSliceExcept return a new slice that without the element appears in the
// second slice.
func StringSliceExcept(slice []string, except []string) []string {
	newSlice := []string{}

	for _, c := range slice {
		if pos := strAt(except, c); pos == -1 {
			newSlice = append(newSlice, c)
		}
	}
	return newSlice
}

// StringSliceContainAny return true iff the container contain one of the
// element at target slice. If the slice is empty, it will return false.
func StringSliceContainAny(container []string, slice []string) bool {
	if len(slice) == 0 {
		return false
	}
	for _, s := range slice {
		if pos := strAt(container, s); pos != -1 {
			return true
		}
	}
	return false
}

// StringSliceContainAll return true iff the container contain all elements of
// the target. If the target slice is empty, it will return true.
func StringSliceContainAll(container []string, slice []string) bool {
	if len(container) < len(slice) {
		return false
	}
	for _, s := range slice {
		if pos := strAt(container, s); pos == -1 {
			return false
		}
	}
	return true
}

func strAt(slice []string, str string) int {
	for pos, s := range slice {
		if s == str {
			return pos
		}
	}
	return -1
}
