package protocol

import (
	"sort"
	"strings"
)

type ResponseType struct {
	orderedList []string
	Raw         string
}

func (rt ResponseType) Equal(other ResponseType) bool {
	if len(rt.orderedList) != len(other.orderedList) {
		return false
	}
	for idx, typ := range rt.orderedList {
		if typ != other.orderedList[idx] {
			return false
		}
	}
	return true
}

func ParseResponseType(str string) ResponseType {
	list := parseSpaceDelimitedString(str)
	sort.Strings(list)
	return ResponseType{orderedList: list, Raw: str}
}

func NewResponseType(responseTypes []string) ResponseType {
	list := make([]string, len(responseTypes))
	_ = copy(list, responseTypes)
	sort.Strings(list)
	return ResponseType{orderedList: list, Raw: strings.Join(responseTypes, " ")}
}
