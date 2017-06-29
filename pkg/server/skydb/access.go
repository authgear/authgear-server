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
	"fmt"
	"sort"
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

	// DiscoverOrCompareFieldAccessMode means the access mode is for discovery
	// or compare
	DiscoverOrCompareFieldAccessMode

	// CompareFieldAccessMode means the access mode is for query
	CompareFieldAccessMode
)

// FieldACL contains all field ACL rules for all record types
type FieldACL struct {
	recordTypes map[string]FieldACLEntryList
}

// NewFieldACL returns a struct of FieldACL with a list of field ACL entries.
func NewFieldACL(list FieldACLEntryList) FieldACL {
	acl := FieldACL{
		recordTypes: map[string]FieldACLEntryList{},
	}

	sort.Sort(list)

	for _, entry := range list {
		perRecordList, _ := acl.recordTypes[entry.RecordType]
		acl.recordTypes[entry.RecordType] = append(perRecordList, entry)
	}

	return acl
}

// AllEntries return all ACL entries in FieldACL.
func (acl FieldACL) AllEntries() FieldACLEntryList {
	var result FieldACLEntryList
	for _, entries := range acl.recordTypes {
		result = append(result, entries...)
	}
	return result
}

// Accessible returns true when the access mode is allowed access
func (acl FieldACL) Accessible(
	recordType string,
	field string,
	mode FieldAccessMode,
	userInfo *UserInfo,
	record *Record,
) bool {
	iter := NewFieldACLIterator(acl, recordType, field)
	for {
		entry := iter.Next()
		if entry == nil {
			// There is no matching ACL entry, the fallback is to grant access.
			return true
		}

		if !entry.UserRole.Match(userInfo, record) {
			continue
		}

		return entry.Accessible(mode)
	}
}

type FieldACLIterator struct {
	acl         FieldACL
	recordType  string
	recordField string

	nextRecordTypes []string
	nextEntries     []FieldACLEntry
	eof             bool
}

func NewFieldACLIterator(acl FieldACL, recordType, recordField string) *FieldACLIterator {
	return &FieldACLIterator{
		acl:             acl,
		recordType:      recordType,
		recordField:     recordField,
		nextRecordTypes: []string{recordType, WildcardRecordType},
	}
}

func (i *FieldACLIterator) Next() *FieldACLEntry {
	if i.eof {
		return nil
	}

	var nextEntry FieldACLEntry
	for {
		for len(i.nextEntries) == 0 {
			if len(i.nextRecordTypes) == 0 {
				i.eof = true
				return nil
			}

			var nextRecordType string
			nextRecordType, i.nextRecordTypes = i.nextRecordTypes[0], i.nextRecordTypes[1:]
			i.nextEntries, _ = i.acl.recordTypes[nextRecordType]
		}

		nextEntry, i.nextEntries = i.nextEntries[0], i.nextEntries[1:]
		if (nextEntry.RecordType == WildcardRecordType || nextEntry.RecordType == i.recordType) &&
			(nextEntry.RecordField == WildcardRecordField || nextEntry.RecordField == i.recordField) {
			break
		}
	}
	return &nextEntry
}

// FieldACLEntryList contains a list of field ACL entries
type FieldACLEntryList []FieldACLEntry

func (list FieldACLEntryList) Len() int           { return len(list) }
func (list FieldACLEntryList) Swap(i, j int)      { list[i], list[j] = list[j], list[i] }
func (list FieldACLEntryList) Less(i, j int) bool { return list[i].Compare(list[j]) < 0 }

// FieldACLEntry contains a single field ACL entry
type FieldACLEntry struct {
	RecordType   string
	RecordField  string
	UserRole     FieldUserRole
	Writable     bool
	Readable     bool
	Comparable   bool
	Discoverable bool
}

// Compare the order for evaluation with the other entry.
//
// This function returns negative when the specified entry have a lower
// priority.
func (entry FieldACLEntry) Compare(other FieldACLEntry) int {
	compare := func(a, b, wildcard string) int {
		if a == wildcard && b != wildcard {
			return 1
		} else if b == wildcard && a != wildcard {
			return -1
		}
		return strings.Compare(a, b)
	}

	result := compare(entry.RecordType, other.RecordType, WildcardRecordType)
	if result != 0 {
		return result
	}

	result = compare(entry.RecordField, other.RecordField, WildcardRecordField)
	if result != 0 {
		return result
	}

	return entry.UserRole.Compare(other.UserRole)
}

// Accessible returns true when the entry grants access for
// the specified access mode. This function does not consider whether
// the entry matches the user role or record type.
func (entry FieldACLEntry) Accessible(mode FieldAccessMode) bool {
	return (mode == ReadFieldAccessMode && entry.Readable) ||
		(mode == WriteFieldAccessMode && entry.Writable) ||
		(mode == CompareFieldAccessMode && entry.Comparable) ||
		(mode == DiscoverOrCompareFieldAccessMode && (entry.Discoverable || entry.Comparable))
}

// FieldUserRoleType denotes the type of field user role, which
// specify who can access certain fields.
type FieldUserRoleType string

const (
	// OwnerFieldUserRoleType means field is accessible by the record owner.
	OwnerFieldUserRoleType FieldUserRoleType = "_owner"

	// SpecificUserFieldUserRoleType means field is accessible by a specific user.
	SpecificUserFieldUserRoleType = "_user_id"

	// DynamicUserFieldUserRoleType means field is accessible by user contained in another field.
	DynamicUserFieldUserRoleType = "_field"

	// DefinedRoleFieldUserRoleType means field is accessible by a users of specific role.
	DefinedRoleFieldUserRoleType = "_role"

	// AnyUserFieldUserRoleType means field is accessible by any authenticated user.
	AnyUserFieldUserRoleType = "_any_user"

	// PublicFieldUserRoleType means field is accessible by public.
	PublicFieldUserRoleType = "_public"
)

// Compare compares two user role type in the order of evaluation.
func (userRoleType FieldUserRoleType) Compare(other FieldUserRoleType) int {
	if userRoleType == other {
		return 0
	}

	for _, eachType := range []FieldUserRoleType{
		OwnerFieldUserRoleType,
		SpecificUserFieldUserRoleType,
		DynamicUserFieldUserRoleType,
		DefinedRoleFieldUserRoleType,
		AnyUserFieldUserRoleType,
		PublicFieldUserRoleType,
	} {
		if userRoleType == eachType {
			return -1
		} else if other == eachType {
			return 1
		}
	}
	return 0
}

// FieldUserRole contains field user role information and checks whether
// a user matches the user role.
type FieldUserRole struct {
	// Type contains the type of the user role.
	Type FieldUserRoleType

	// Data is information specific to the type of user role.
	Data string
}

// NewFieldUserRole returns a FieldUserRole struct from the user role
// specification.
func NewFieldUserRole(roleString string) FieldUserRole {
	components := strings.SplitN(roleString, ":", 2)
	roleType := FieldUserRoleType(components[0])
	switch roleType {
	case OwnerFieldUserRoleType, AnyUserFieldUserRoleType, PublicFieldUserRoleType:
		if len(components) > 1 {
			panic(fmt.Sprintf(`unexpected user role string "%s"`, roleString))
		}
		return FieldUserRole{roleType, ""}
	case SpecificUserFieldUserRoleType, DynamicUserFieldUserRoleType, DefinedRoleFieldUserRoleType:
		if len(components) != 2 {
			panic(fmt.Sprintf(`unexpected user role string "%s"`, roleString))
		}
		return FieldUserRole{roleType, components[1]}
	default:
		panic(fmt.Sprintf(`unexpected user role string "%s"`, roleString))

	}
}

// String returns the user role specification in string representation.
func (r FieldUserRole) String() string {
	switch r.Type {
	case OwnerFieldUserRoleType, AnyUserFieldUserRoleType, PublicFieldUserRoleType:
		return string(r.Type)
	case SpecificUserFieldUserRoleType, DynamicUserFieldUserRoleType, DefinedRoleFieldUserRoleType:
		return fmt.Sprintf("%s:%s", r.Type, r.Data)
	default:
		panic(fmt.Sprintf(`unexpected field user role type "%s"`, r.Type))
	}
}

// Compare compares two FieldUserRole according to the order of evaluation.
func (r FieldUserRole) Compare(other FieldUserRole) int {
	result := r.Type.Compare(other.Type)
	if result != 0 {
		return result
	}
	return strings.Compare(r.Data, other.Data)
}

// Match returns true if the specifid UserInfo and Record matches the
// user role.
func (r FieldUserRole) Match(userinfo *UserInfo, record *Record) bool {
	if r.Type == PublicFieldUserRoleType {
		return true
	}

	// All the other types requires UserInfo
	if userinfo == nil {
		return false
	}

	switch r.Type {
	case OwnerFieldUserRoleType:
		// TODO
		return false
	case SpecificUserFieldUserRoleType:
		return userinfo.ID == r.Data
	case DynamicUserFieldUserRoleType:
		// TODO
		return false
	case DefinedRoleFieldUserRoleType:
		for _, role := range userinfo.Roles {
			if role == r.Data {
				return true
			}
		}
		return false
	case AnyUserFieldUserRoleType:
		return true
	default:
		panic(fmt.Sprintf(`unexpected field user role type "%s"`, r.Type))
	}
}

var defaultFieldUserRole = FieldUserRole{PublicFieldUserRoleType, ""}

// WildcardRecordType is a special record type that applies to all record types
const WildcardRecordType = "*"

// WildcardRecordField is a special record field that applies to all record fields
const WildcardRecordField = "*"
