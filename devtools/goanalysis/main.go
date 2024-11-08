package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/authgear/authgear-server/devtools/goanalysis/pkg/httpclient"
)

func main() {
	multichecker.Main(httpclient.Analyzer)
}
