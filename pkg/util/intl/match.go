package intl

import (
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"golang.org/x/text/language"
)

var matcherCache = cache.New(5*time.Minute, 10*time.Minute)

func toLanguageTags(ss []string) []language.Tag {
	tags := make([]language.Tag, len(ss))
	for i, item := range ss {
		tags[i] = language.Make(item)
	}
	return tags
}

func getMatcherKey(supported []language.Tag) string {
	ss := make([]string, len(supported))
	for i, item := range supported {
		ss[i] = item.String()
	}
	return strings.Join(ss, ",")
}

func GetMatcher(supported []language.Tag) language.Matcher {
	key := getMatcherKey(supported)
	matcherIface, ok := matcherCache.Get(key)
	if ok {
		return matcherIface.(language.Matcher)
	}

	// pprof shows golang.org/x/text@0.1{3,4}.0 language.NewMatcher is a CPU hot spot.
	// So we cache intances of matcher.
	matcher := language.NewMatcher(supported)
	matcherCache.Set(key, matcher, cache.DefaultExpiration)
	return matcher
}

// Match matches preferredLanguageTags to supportedLanguageTags
// using fallbackLanguageTag as fallback.
// NOTE(tung): Replaced by BestMatch. Use BestMatch instead.
// This function were keep just for reference and testing
func Match_Deprecated(preferredLanguageTags []string, supportedLanguageTags SupportedLanguages) (int, language.Tag) {
	if len(supportedLanguageTags) <= 0 {
		return -1, language.Und
	}

	supported := toLanguageTags(supportedLanguageTags)
	preferred := toLanguageTags(preferredLanguageTags)

	matcher := GetMatcher(supported)

	_, idx, _ := matcher.Match(preferred...)
	tag := supported[idx]

	return idx, tag
}

// matcher.Match will not return tags with higher confidence
// For example, with supported tags zh-CN, zh-HK
// matcher.Match("zh-Hant") will return zh-CN, which confidence is Low,
// but not zh-HK which confidence is High.
// This function is an implementation of Match trying to return an option with higher confidence.
func BestMatch(preferredLanguageTags []string, supportedLanguageTags SupportedLanguages) (int, language.Tag) {
	if len(supportedLanguageTags) <= 0 {
		return -1, language.Und
	}

	supportedTags := toLanguageTags(supportedLanguageTags)
	preferredTags := toLanguageTags(preferredLanguageTags)

	if len(preferredTags) <= 0 {
		return 0, supportedTags[0]
	}

	var selectedTagIdx int = -1
	var selectedTagConfidence language.Confidence
	for _, pt := range preferredTags {
		preferredTag := pt

		for idx, t := range supportedTags {
			supportedTag := t
			matcher := GetMatcher([]language.Tag{supportedTag})
			_, _, confidence := matcher.Match(preferredTag)

			// If exact match, choose this option without considering others
			if confidence == language.Exact {
				selectedTagIdx = idx
				selectedTagConfidence = confidence
				break
			}

			// Else, select the option with highest confidence
			if confidence > selectedTagConfidence || selectedTagIdx == -1 {
				selectedTagIdx = idx
				selectedTagConfidence = confidence
			}
		}

		// If confidence is not No, use this match as the result and do not look at other preferred tags
		if selectedTagConfidence != language.No {
			break
		}
	}

	tag := supportedTags[selectedTagIdx]

	return selectedTagIdx, tag
}
