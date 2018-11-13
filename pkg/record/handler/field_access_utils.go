package handler

import (
	"sort"

	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
)

type FieldAccessResponse struct {
	Access record.FieldACLEntryList `json:"access"`
}

func NewFieldAccessResponse(fieldACL record.FieldACL) FieldAccessResponse {
	fieldACLEntries := fieldACL.AllEntries()
	if len(fieldACLEntries) == 0 {
		// Make sure the response contains array with 0 items rather than nil.
		fieldACLEntries = make(record.FieldACLEntryList, 0)
	} else {
		sort.Sort(fieldACLEntries)
	}

	return FieldAccessResponse{
		Access: fieldACLEntries,
	}
}
