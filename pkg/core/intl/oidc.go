package intl

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/text/language"
)

// LocalizeJSONObject returns the localized value of key in jsonObject according to preferredLanguageTags.
func LocalizeJSONObject(preferredLanguageTags []string, jsonObject map[string]interface{}, key string) string {
	prefix := fmt.Sprintf("%s#", key)
	m := map[string]string{}
	for k, v := range jsonObject {
		stringValue, ok := v.(string)
		if !ok {
			continue
		}
		if k == key {
			m[""] = stringValue
		} else if strings.HasPrefix(k, prefix) {
			tag := strings.TrimPrefix(k, prefix)
			m[tag] = stringValue
		}
	}
	_, value := Localize(preferredLanguageTags, m)
	return value
}

// LocalizeStringMap returns the localized value of key in stringMap according to preferredLanguageTags.
func LocalizeStringMap(preferredLanguageTags []string, stringMap map[string]string, key string) string {
	prefix := fmt.Sprintf("%s#", key)
	m := map[string]string{}
	for k, stringValue := range stringMap {
		if k == key {
			m[""] = stringValue
		} else if strings.HasPrefix(k, prefix) {
			tag := strings.TrimPrefix(k, prefix)
			m[tag] = stringValue
		}
	}
	_, value := Localize(preferredLanguageTags, m)
	return value
}

// Localize selects the best value from m according to preferredLanguageTags.
// m must be non-empty.
func Localize(preferredLanguageTags []string, m map[string]string) (language.Tag, string) {
	if len(m) <= 0 {
		return language.Und, ""
	}

	supportedTagStrings := make([]string, len(m))
	for tagStr := range m {
		supportedTagStrings = append(supportedTagStrings, tagStr)
	}

	// The first item in tags is used as fallback.
	// So we have sort the templates so that template with empty
	// language tag comes first.
	sort.Slice(supportedTagStrings, func(i, j int) bool {
		return supportedTagStrings[i] < supportedTagStrings[j]
	})

	supportedTags := make([]language.Tag, len(supportedTagStrings))
	for i, item := range supportedTagStrings {
		supportedTags[i] = language.Make(item)
	}
	matcher := language.NewMatcher(supportedTags)

	preferredTags := make([]language.Tag, len(preferredLanguageTags))
	for i, tagStr := range preferredLanguageTags {
		preferredTags[i] = language.Make(tagStr)
	}

	_, idx, _ := matcher.Match(preferredTags...)

	tag := supportedTags[idx]
	pattern := m[supportedTagStrings[idx]]

	return tag, pattern
}
