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
	"strings"
)

// RecordACLEntry grants access to a record by relation or by user_id
type RecordACLEntry struct {
	Relation string         `json:"relation,omitempty"`
	Role     string         `json:"role,omitempty"`
	Level    RecordACLLevel `json:"level"`
	UserID   string         `json:"user_id,omitempty"`
	Public   bool           `json:"public,omitempty"`
}

// RecordACLLevel represent the operation a user granted on a resource
type RecordACLLevel string

// ReadLevel and WriteLevel is self-explanatory
const (
	ReadLevel   RecordACLLevel = "read"
	WriteLevel                 = "write"
	CreateLevel                = "create"
)

// NewRecordACLEntryRelation returns an ACE on relation
func NewRecordACLEntryRelation(relation string, level RecordACLLevel) RecordACLEntry {
	return RecordACLEntry{
		Relation: relation,
		Level:    level,
	}
}

// NewRecordACLEntryDirect returns an ACE for a specific user
func NewRecordACLEntryDirect(userID string, level RecordACLLevel) RecordACLEntry {
	return RecordACLEntry{
		Relation: "$direct",
		Level:    level,
		UserID:   userID,
	}
}

// NewRecordACLEntryRole return an ACE on role
func NewRecordACLEntryRole(role string, level RecordACLLevel) RecordACLEntry {
	return RecordACLEntry{
		Role:  role,
		Level: level,
	}
}

// NewRecordACLEntryPublic return an ACE on public access
func NewRecordACLEntryPublic(level RecordACLLevel) RecordACLEntry {
	return RecordACLEntry{
		Public: true,
		Level:  level,
	}
}

func (ace *RecordACLEntry) Accessible(userinfo *UserInfo, level RecordACLLevel) bool {
	if ace.Public {
		return ace.AccessibleLevel(level)
	}
	if userinfo == nil {
		return false
	}
	if userinfo.ID == ace.UserID {
		if ace.AccessibleLevel(level) {
			return true
		}
	}
	for _, role := range userinfo.Roles {
		if role == ace.Role {
			if ace.AccessibleLevel(level) {
				return true
			}
		}
	}
	return false
}

func (ace *RecordACLEntry) AccessibleLevel(level RecordACLLevel) bool {
	if level == ReadLevel {
		return true
	}
	if level == ace.Level && level == WriteLevel {
		return true
	}
	return false
}

// RecordACL is a list of ACL entries defining access control for a record
type RecordACL []RecordACLEntry

// NewRecordACL returns a new RecordACL
func NewRecordACL(entries []RecordACLEntry) RecordACL {
	acl := make(RecordACL, len(entries))
	for i, v := range entries {
		acl[i] = v
	}
	return acl
}

// Accessible checks whether provided user info has certain access level
func (acl RecordACL) Accessible(userinfo *UserInfo, level RecordACLLevel) bool {
	if len(acl) == 0 {
		// default behavior of empty ACL
		return true
	}

	accessible := false
	for _, ace := range acl {
		if ace.Accessible(userinfo, level) {
			accessible = true
		}
	}

	return accessible
}

// FieldAccessMode is the intended access operation to be granted access
type FieldAccessMode int

const (
	// ReadFieldAccessMode means the access mode is for reading
	ReadFieldAccessMode FieldAccessMode = iota + 1

	// WriteFieldAccessMode means the access mode is for writing
	WriteFieldAccessMode

	// DiscoverFieldAccessMode means the access mode is for discovery
	DiscoverFieldAccessMode

	// CompareFieldAccessMode means the access mode is for query
	CompareFieldAccessMode
)

// FieldACL contains all field ACL rules for all record types
type FieldACL struct {
	wildcardRecordType FieldACLEntryList
	recordTypes        map[string]FieldACLEntryList
}

// NewFieldACL returns a struct of FieldACL with a list of field ACL entries.
func NewFieldACL(list FieldACLEntryList) FieldACL {
	acl := FieldACL{
		wildcardRecordType: FieldACLEntryList{},
		recordTypes:        map[string]FieldACLEntryList{},
	}

	for _, entry := range list {
		if entry.RecordType == WildcardRecordType {
			acl.wildcardRecordType = append(acl.wildcardRecordType, entry)
			continue
		}

		perRecordList, _ := acl.recordTypes[entry.RecordType]
		acl.recordTypes[entry.RecordType] = append(perRecordList, entry)
	}

	return acl
}

// NewFieldACLDefault returns a struct of FieldACL with a default setting
// if the default setting is not otherwise specified.
func NewFieldACLDefault(list FieldACLEntryList, defEntry FieldACLEntry) FieldACL {
	acl := NewFieldACL(list)
	entry := acl.FindDefaultEntry()
	if entry == nil {
		// There is no entry with wildcard record type and record field,
		// add the default entry to the wildcardRecordType list
		defEntry.RecordType = WildcardRecordType
		defEntry.RecordField = WildcardRecordField
		defEntry.UserRole = "_public"
		acl.wildcardRecordType = append(acl.wildcardRecordType, defEntry)
	}

	return acl
}

// AllEntries return all ACL entries in FieldACL.
func (acl FieldACL) AllEntries() FieldACLEntryList {
	result := acl.wildcardRecordType
	for _, entries := range acl.recordTypes {
		result = append(result, entries...)
	}
	return result
}

// FindDefaultEntry finds the default ACL entry in FieldACL.
//
// This function returns nil if the default ACL entry is not contained
// in the FieldACL.
func (acl FieldACL) FindDefaultEntry() *FieldACLEntry {
	return acl.wildcardRecordType.findDefaultEntry()
}

// Accessible returns true when the access mode is allowed access
func (acl FieldACL) Accessible(
	userinfo *UserInfo,
	record *Record,
	recordType string,
	field string,
	mode FieldAccessMode,
) bool {
	if acl.wildcardRecordType.Accessible(userinfo, record, recordType, field, mode) {
		return true
	}
	if list, ok := acl.recordTypes[recordType]; ok {
		return list.Accessible(userinfo, record, recordType, field, mode)
	}
	return false
}

// FieldACLEntryList contains a list of field ACL entries
type FieldACLEntryList []FieldACLEntry

// Accessible returns true when the access mode is allowed access
func (list FieldACLEntryList) Accessible(
	userinfo *UserInfo,
	record *Record,
	recordType string,
	field string,
	mode FieldAccessMode,
) bool {
	for _, entry := range list {
		if entry.Accessible(userinfo, record, recordType, field, mode) {
			return true
		}
	}
	return false
}

func (list FieldACLEntryList) findDefaultEntry() *FieldACLEntry {
	for _, entry := range list {
		if entry.RecordType == WildcardRecordType &&
			entry.RecordField == WildcardRecordField &&
			entry.UserRole == "_public" {
			return &entry
		}
	}
	return nil
}

func (list FieldACLEntryList) Len() int      { return len(list) }
func (list FieldACLEntryList) Swap(i, j int) { list[i], list[j] = list[j], list[i] }
func (list FieldACLEntryList) Less(i, j int) bool {
	// compare is similar to strings.Compare except that specified wildcard
	// string will be less than non-wildcard string.
	compare := func(a, b, wildcard string) int {
		if a == wildcard && b != wildcard {
			return -1
		} else if b == wildcard && a != wildcard {
			return 1
		}
		return strings.Compare(a, b)
	}

	result := compare(list[i].RecordType, list[j].RecordType, WildcardRecordType)
	if result != 0 {
		return result < 0
	}

	result = compare(list[i].RecordField, list[j].RecordField, WildcardRecordField)
	if result != 0 {
		return result < 0
	}

	return strings.Compare(list[i].UserRole, list[j].UserRole) < 0
}

// FieldACLEntry contains a single field ACL entry
type FieldACLEntry struct {
	RecordType   string
	RecordField  string
	UserRole     string
	Writable     bool
	Readable     bool
	Comparable   bool
	Discoverable bool
}

// Accessible returns true when the access mode is allowed access
func (entry FieldACLEntry) Accessible(
	userinfo *UserInfo,
	record *Record,
	recordType string,
	field string,
	mode FieldAccessMode,
) bool {
	if (entry.RecordType != recordType && entry.RecordType != WildcardRecordType) ||
		(entry.RecordField != field && entry.RecordField != WildcardRecordField) {
		return false
	}

	// TODO: Check access here

	return (mode == ReadFieldAccessMode && entry.Readable) ||
		(mode == WriteFieldAccessMode && entry.Writable) ||
		(mode == CompareFieldAccessMode && entry.Comparable) ||
		(mode == DiscoverFieldAccessMode && entry.Discoverable)
}

// WildcardRecordType is a special record type that applies to all record types
const WildcardRecordType = "*"

// WildcardRecordField is a special record field that applies to all record fields
const WildcardRecordField = "*"
