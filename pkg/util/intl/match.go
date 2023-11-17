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
func Match(preferredLanguageTags []string, supportedLanguageTags SupportedLanguages) (int, language.Tag) {
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
