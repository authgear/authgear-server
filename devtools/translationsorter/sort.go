package main

import (
	"sort"
	"strings"
)

func SortKeyValuePairs(kvPairs []KeyValuePair) []KeyValuePair {
	out := make([]KeyValuePair, len(kvPairs))
	copy(out, kvPairs)

	sort.SliceStable(out, func(i, j int) bool {
		return cmpKey(out[i].Key, out[j].Key)
	})
	return out
}

func isV2Key(key string) bool {
	return strings.HasPrefix(key, "v2")
}

func cmpKey(a, b string) bool {
	switch {
	case isV2Key(a) && !isV2Key(b):
		return false
	case !isV2Key(a) && isV2Key(b):
		return true
		// We want to sort V2 keys ONLY, v1 keys orders should be preserved
	case !isV2Key(a) && !isV2Key(b):
		return false
	default:
		return a < b
	}
}
