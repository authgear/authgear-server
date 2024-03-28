package transport

import (
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/geoip"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func ConfigureGeoipRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("GET").WithPathPattern("/api/geoip")
}

type GeoipHandler struct {
}

type GeoipInfo struct {
	CountryCode string `json:"country_code"`
}

func (h *GeoipHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	makeGeoipInfo := func() ([]byte, error) {
		requestIP := httputil.GetIP(r, false)
		geoipInfo, ok := geoip.DefaultDatabase.IPString(requestIP)
		if !ok {
			return nil, nil
		}

		b, err := json.Marshal(&GeoipInfo{
			CountryCode: geoipInfo.CountryCode,
		})
		if err != nil {
			return nil, err
		}

		return b, nil
	}

	resp, err := makeGeoipInfo()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
