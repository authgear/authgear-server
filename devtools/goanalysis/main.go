package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/authgear/authgear-server/devtools/goanalysis/pkg/contextbackground"
	"github.com/authgear/authgear-server/devtools/goanalysis/pkg/httpclient"
	"github.com/authgear/authgear-server/devtools/goanalysis/pkg/requestcontext"
	"github.com/authgear/authgear-server/devtools/goanalysis/pkg/timeunixutc"
	"github.com/authgear/authgear-server/devtools/goanalysis/pkg/util/vettedposutil"
	"github.com/authgear/authgear-server/devtools/goanalysis/pkg/vettedpos"
)

func main() {
	pos, err := vettedposutil.NewVettedPositionsFromFile("./.vettedpositions")
	if err != nil {
		panic(err)
	}

	// multichecker.Main calls os.Exit internally.
	// So the statements after it, or any defer statement before it will not run.
	// To report unused vetted position,
	// we need to register a new Analyzer.

	// Note that the Requires of vettedpos.Analyzer appears twice.
	// This is due to my observation that the Requires of a analyzer does not
	// contribute to the final report.
	// Only analyzer that is passed to multichecker will be included in the final report.
	contextbackgroundAnalyzer := contextbackground.NewAnalyzer(pos)
	requestcontetAnalyzer := requestcontext.NewAnalyzer(pos)
	vettedposAnalzyer := vettedpos.NewAnalyzer(pos,
		contextbackgroundAnalyzer,
		requestcontetAnalyzer,
	)

	multichecker.Main(
		httpclient.Analyzer,
		timeunixutc.Analyzer,
		contextbackgroundAnalyzer,
		requestcontetAnalyzer,
		vettedposAnalzyer,
	)
}
