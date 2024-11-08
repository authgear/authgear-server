package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/authgear/authgear-server/devtools/goanalysis/pkg/httpclient"
	"github.com/authgear/authgear-server/devtools/goanalysis/pkg/timeunixutc"
)

func main() {
	multichecker.Main(httpclient.Analyzer, timeunixutc.Analyzer)
}