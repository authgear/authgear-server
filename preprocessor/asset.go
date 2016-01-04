package preprocessor

import (
	"net/http"

	"github.com/oursky/skygear/asset"
	"github.com/oursky/skygear/router"
)

type AssetStorePreprocessor struct {
	Store asset.Store
}

func (p AssetStorePreprocessor) Preprocess(payload *router.Payload, response *router.Response) int {
	payload.AssetStore = p.Store
	return http.StatusOK
}
