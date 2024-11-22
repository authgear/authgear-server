package requestcontext

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/authgear/authgear-server/devtools/goanalysis/pkg/util/vettedposutil"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), NewAnalyzer(vettedposutil.NewEmptyVettedPositions()), "basic")
}
