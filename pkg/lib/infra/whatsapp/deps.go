package whatsapp

import (
	"net/http"
	"time"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type HTTPClient struct {
	*http.Client
}

func NewHTTPClient() HTTPClient {
	return HTTPClient{
		httputil.NewExternalClient(5 * time.Second),
	}
}

var DependencySet = wire.NewSet(
	NewServiceLogger,
	NewHTTPClient,
	NewWhatsappOnPremisesClient,
	NewWhatsappCloudAPIClient,
	wire.Struct(new(TokenStore), "*"),
	wire.Struct(new(Service), "*"),
)
