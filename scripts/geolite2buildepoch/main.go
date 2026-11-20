package main

// Prints the build time of a GeoLite2 mmdb file in RFC3339, as asserted in pkg/util/geoip/geoip_test.go.
// Usage: go run ./scripts/geolite2buildepoch <path-to-mmdb>

import (
	"fmt"
	"os"
	"time"

	"github.com/oschwald/geoip2-golang"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <path-to-mmdb>\n", os.Args[0])
		os.Exit(1)
	}

	reader, err := geoip2.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open mmdb: %v\n", err)
		os.Exit(1)
	}

	//nolint:gosec // G115
	sec := int64(reader.Metadata().BuildEpoch)
	if err := reader.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to close mmdb: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(time.Unix(sec, 0).UTC().Format(time.RFC3339))
}
