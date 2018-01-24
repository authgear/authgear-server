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
	"encoding/json"
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
	WriteLevel  RecordACLLevel = "write"
	CreateLevel RecordACLLevel = "create"
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

func (ace *RecordACLEntry) Accessible(authinfo *AuthInfo, level RecordACLLevel) bool {
	if ace.Public {
		return ace.AccessibleLevel(level)
	}
	if authinfo == nil {
		return false
	}
	if authinfo.ID == ace.UserID {
		if ace.AccessibleLevel(level) {
			return true
		}
	}
	for _, role := range authinfo.Roles {
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
	if level == ace.Level && level == CreateLevel {
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
func (acl RecordACL) Accessible(authinfo *AuthInfo, level RecordACLLevel) bool {
	if len(acl) == 0 {
		// default behavior of empty ACL
		return true
	}

	accessible := false
	for _, ace := range acl {
		if ace.Accessible(authinfo, level) {
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

// FieldACL contains all field ACL rules for all record types. This struct
// provides functions for evaluating whether access can be granted for a
// request.
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
//
// The accessibility of a field access request is determined by the first
// matching rule. If no matching rule is found, the default rule is to
// grant access.
func (acl FieldACL) Accessible(
	recordType string,
	field string,
	mode FieldAccessMode,
	authInfo *AuthInfo,
	record *Record,
) bool {
	// Create an iterator for Field ACL rules that applies to
	// the specified record type and field. The iterator will handle the
	// scenario where wildcard record type and field.
	iter := NewFieldACLIterator(acl, recordType, field)
	entry := iter.Next()

	// There is no ACL entry that matches the field of record type,
	// the fallback is to grant access
	if entry == nil {
		return true
	}

	for ; entry != nil; entry = iter.Next() {
		if entry.UserRole.Match(authInfo, record) {
			return entry.Accessible(mode)
		}
	}

	// There is no ACL entry that matches the user role,
	// the fallback is access denied
	return false
}

// FieldACLIterator iterates FieldACL to find a list of rules that apply
// to the specified record type and record field.
//
// The iterator does not consider the access mode, the AuthInfo and the Record
// of individual access. So the result is always the same as long as the
// FieldACL setting is unchanged. The list of rules can then be considered
// one by one, which is specific to each individual request.
type FieldACLIterator struct {
	acl         FieldACL
	recordType  string
	recordField string

	nextRecordTypes []string
	nextEntries     []FieldACLEntry
	eof             bool
}

// NewFieldACLIterator creates a new iterator.
func NewFieldACLIterator(acl FieldACL, recordType, recordField string) *FieldACLIterator {
	return &FieldACLIterator{
		acl:             acl,
		recordType:      recordType,
		recordField:     recordField,
		nextRecordTypes: []string{recordType, WildcardRecordType},
	}
}

// Next returns the next FieldACLEntry. If there is no more entries to return,
// this function will return nil.
func (i *FieldACLIterator) Next() *FieldACLEntry {
	if i.eof {
		return nil
	}

	var nextEntry FieldACLEntry
	for {
		// Always populate the nextEntries var with the upcoming entries.
		// If there is no more entries to populate, the iterator will stop and
		// always return nil.
		for len(i.nextEntries) == 0 {
			if len(i.nextRecordTypes) == 0 {
				i.eof = true
				return nil
			}

			var nextRecordType string
			nextRecordType, i.nextRecordTypes = i.nextRecordTypes[0], i.nextRecordTypes[1:]
			i.nextEntries, _ = i.acl.recordTypes[nextRecordType]
		}

		// Get next entry.
		nextEntry, i.nextEntries = i.nextEntries[0], i.nextEntries[1:]
		if nextEntry.RecordField == i.recordField || nextEntry.RecordField == WildcardRecordField {
			// Stop the iterator in the next iteration when
			// - no more entries for the record type
			// - two entries have different field
			i.eof = len(i.nextEntries) == 0 ||
				nextEntry.RecordField != i.nextEntries[0].RecordField
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
	RecordType   string        `json:"record_type"`
	RecordField  string        `json:"record_field"`
	UserRole     FieldUserRole `json:"user_role"`
	Writable     bool          `json:"writable"`
	Readable     bool          `json:"readable"`
	Comparable   bool          `json:"comparable"`
	Discoverable bool          `json:"discoverable"`
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
	SpecificUserFieldUserRoleType FieldUserRoleType = "_user_id"

	// DynamicUserFieldUserRoleType means field is accessible by user contained in another field.
	DynamicUserFieldUserRoleType FieldUserRoleType = "_field"

	// DefinedRoleFieldUserRoleType means field is accessible by a users of specific role.
	DefinedRoleFieldUserRoleType FieldUserRoleType = "_role"

	// AnyUserFieldUserRoleType means field is accessible by any authenticated user.
	AnyUserFieldUserRoleType FieldUserRoleType = "_any_user"

	// PublicFieldUserRoleType means field is accessible by public.
	PublicFieldUserRoleType FieldUserRoleType = "_public"
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

// RecordDependent returns true if this user role type requires
// record data when evaluating accessibility.
func (userRoleType FieldUserRoleType) RecordDependent() bool {
	return userRoleType == OwnerFieldUserRoleType ||
		userRoleType == DynamicUserFieldUserRoleType
}

// FieldUserRole contains field user role information and checks whether
// a user matches the user role.
type FieldUserRole struct {
	// Type contains the type of the user role.
	Type FieldUserRoleType

	// Data is information specific to the type of user role.
	Data string
}

// ParseFieldUserRole parses a user role string to a FieldUserRole.
func ParseFieldUserRole(roleString string) (FieldUserRole, error) {
	components := strings.SplitN(roleString, ":", 2)
	roleType := FieldUserRoleType(components[0])
	switch roleType {
	case OwnerFieldUserRoleType, AnyUserFieldUserRoleType, PublicFieldUserRoleType:
		if len(components) > 1 {
			return FieldUserRole{}, fmt.Errorf(`unexpected user role string "%s"`, roleString)
		}
		return FieldUserRole{roleType, ""}, nil
	case SpecificUserFieldUserRoleType, DynamicUserFieldUserRoleType, DefinedRoleFieldUserRoleType:
		if len(components) != 2 {
			return FieldUserRole{}, fmt.Errorf(`unexpected user role string "%s"`, roleString)
		}
		return FieldUserRole{roleType, components[1]}, nil
	default:
		return FieldUserRole{}, fmt.Errorf(`unexpected user role string "%s"`, roleString)

	}
}

// NewFieldUserRole returns a FieldUserRole struct from the user role
// specification.
func NewFieldUserRole(roleString string) FieldUserRole {
	userRole, err := ParseFieldUserRole(roleString)
	if err != nil {
		panic(err)
	}
	return userRole
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

// Match returns true if the specifid AuthInfo and Record matches the
// user role.
//
// If the specified user role type is record dependent, this function
// returns false if Record is nil.
func (r FieldUserRole) Match(authinfo *AuthInfo, record *Record) bool {
	if r.Type == PublicFieldUserRoleType {
		return true
	}

	// Exit early if authinfo and record is nil
	if authinfo == nil || (r.Type.RecordDependent() && record == nil) {
		return false
	}

	switch r.Type {
	case OwnerFieldUserRoleType:
		return record.OwnerID == authinfo.ID
	case SpecificUserFieldUserRoleType:
		return authinfo.ID == r.Data
	case DynamicUserFieldUserRoleType:
		return r.matchDynamic(authinfo, record)
	case DefinedRoleFieldUserRoleType:
		for _, role := range authinfo.Roles {
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

// matchDynamic is a helper function for returning whether the
// field user role with dynamic user field type matches the specified
// AuthInfo and Record.
func (r FieldUserRole) matchDynamic(authInfo *AuthInfo, record *Record) bool {
	dynamicFieldName := r.Data
	switch fieldVal := record.Get(dynamicFieldName).(type) {
	case string:
		return authInfo.ID == fieldVal
	case []interface{}:
		for _, item := range fieldVal {
			if userID, ok := item.(string); ok && userID == authInfo.ID {
				return true
			}
		}
		return false
	}
	return false
}

// MarshalJSON implements json.Marshaler
func (r *FieldUserRole) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

// UnmarshalJSON implements json.Unmarshaler
func (r *FieldUserRole) UnmarshalJSON(data []byte) (err error) {
	var strValue string
	if err = json.Unmarshal(data, &strValue); err != nil {
		return
	}
	newRole := NewFieldUserRole(strValue)
	*r = newRole
	return
}

var defaultFieldUserRole = FieldUserRole{PublicFieldUserRoleType, ""}

// WildcardRecordType is a special record type that applies to all record types
const WildcardRecordType = "*"

// WildcardRecordField is a special record field that applies to all record fields
const WildcardRecordField = "*"
