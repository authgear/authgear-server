// SPDX-License-Identifier: Apache-2.0

// Original Work
// Copyright 2019 Google LLC.
// https://www.chromium.org/updates/same-site/incompatible-clients

// Modified Work
// Copyright 2020 Oursky Ltd.

// Don’t send `SameSite=None` to known incompatible clients.

package httputil

import (
	"regexp"
	"strconv"
)

var iosVersionRegex = regexp.MustCompile(`\(iP.+; CPU .*OS (\d+)[_\d]*.*\) AppleWebKit\/`)
var macosxVersionRegex = regexp.MustCompile(`\(Macintosh;.*Mac OS X (\d+)_(\d+)[_\d]*.*\) AppleWebKit\/`)
var safariRegex = regexp.MustCompile(`Version\/.* Safari\/`)
var isMacEmbeddedBrowserRegex = regexp.MustCompile(`^Mozilla\/[\.\d]+ \(Macintosh;.*Mac OS X [_\d]+\) AppleWebKit\/[\.\d]+ \(KHTML, like Gecko\)$`)
var isChromiumBasedRegex = regexp.MustCompile(`Chrom(e|ium)`)
var chromiumVersionRegex = regexp.MustCompile(`Chrom[^ \/]+\/(\d+)[\.\d]* `)
var isUcBrowserRegex = regexp.MustCompile(`UCBrowser\/`)
var ucBrowserVersionRegex = regexp.MustCompile(`UCBrowser\/(\d+)\.(\d+)\.(\d+)[\.\d]* `)

func ShouldSendSameSiteNone(useragent string, secure bool) bool {
	// Any cookie with SameSite=None and not Secure will be rejected.
	// So without the Secure attribute, the cookie must NOT have SameSite=None.
	// https://www.chromestatus.com/feature/5633521622188032
	if !secure {
		return false
	}
	return !isSameSiteNoneIncompatible(useragent)
}

// Classes of browsers known to be incompatible.

func isSameSiteNoneIncompatible(useragent string) bool {
	return hasWebKitSameSiteBug(useragent) ||
		dropsUnrecognizedSameSiteCookies(useragent)
}

func hasWebKitSameSiteBug(useragent string) bool {
	return isIosVersion(12, useragent) ||
		(isMacosxVersion(10, 14, useragent) &&
			(isSafari(useragent) || isMacEmbeddedBrowser(useragent)))
}

func dropsUnrecognizedSameSiteCookies(useragent string) bool {
	if isUcBrowser(useragent) {
		return !isUcBrowserVersionAtLeast(12, 13, 2, useragent)
	}

	return isChromiumBased(useragent) &&
		isChromiumVersionAtLeast(51, useragent) &&
		!isChromiumVersionAtLeast(67, useragent)
}

func isIosVersion(major int, useragent string) bool {
	matches := iosVersionRegex.FindStringSubmatch(useragent)
	if len(matches) == 0 {
		return false
	}
	return matches[1] == strconv.Itoa(major)
}

func isMacosxVersion(major int, minor int, useragent string) bool {
	matches := macosxVersionRegex.FindStringSubmatch(useragent)
	if len(matches) == 0 {
		return false
	}
	// Extract digits from first and second capturing groups.
	return matches[1] == strconv.Itoa(major) && matches[2] == strconv.Itoa(minor)
}

func isSafari(useragent string) bool {
	return safariRegex.MatchString(useragent) && !isChromiumBased(useragent)
}

func isMacEmbeddedBrowser(useragent string) bool {
	return isMacEmbeddedBrowserRegex.MatchString(useragent)
}

func isChromiumBased(useragent string) bool {
	return isChromiumBasedRegex.MatchString(useragent)
}

func isChromiumVersionAtLeast(major int, useragent string) bool {
	matches := chromiumVersionRegex.FindStringSubmatch(useragent)
	if len(matches) == 0 {
		return false
	}

	// Extract digits from first capturing group.
	version, _ := strconv.Atoi(matches[1])
	return version >= major

}
func isUcBrowser(useragent string) bool {
	return isUcBrowserRegex.MatchString(useragent)
}

func isUcBrowserVersionAtLeast(major int, minor int, build int, useragent string) bool {
	matches := ucBrowserVersionRegex.FindStringSubmatch(useragent)
	if len(matches) == 0 {
		return false
	}
	// Extract digits from three capturing groups.
	majorVersion, _ := strconv.Atoi(matches[1])
	minorVersion, _ := strconv.Atoi(matches[2])
	buildVersion, _ := strconv.Atoi(matches[3])
	if majorVersion != major {
		return majorVersion > major
	}
	if minorVersion != minor {
		return minorVersion > minor
	}
	return buildVersion >= build
}
