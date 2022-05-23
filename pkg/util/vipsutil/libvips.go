//go:build !authgearlite
// +build !authgearlite

package vipsutil

import (
	"github.com/davidbyttow/govips/v2/vips"
)

// LibvipsInit calls vips_init, and then vips_cache_set_max, etc to
// make libvips available for use.
// The counterpart LibvipsShutdown is intentionally not exposed because
// libvips does not support restart.
// After vips_shutdown has been called, libvips can no longer be used again in the program.
func LibvipsInit() {
	// https://github.com/imgproxy/imgproxy/blob/v3.3.0/vips/vips.go#L45
	// https://github.com/davidbyttow/govips/blob/v2.10.0/vips/govips.go#L59
	cfg := vips.Config{
		ConcurrencyLevel: 1,
		MaxCacheFiles:    0,
		MaxCacheMem:      0,
		MaxCacheSize:     0,
		ReportLeaks:      true,
		CacheTrace:       false,
		CollectStats:     false,
	}
	vips.Startup(&cfg)
	// The default log evel is INFO, which is too noisey.
	vips.LoggingSettings(nil, vips.LogLevelWarning)
}
