//+build wireinject

package loginid

import (
	"net/http"

	"github.com/google/wire"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	interactionflows "github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/flows"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func provideAddLoginIDHandler(
	v *validation.Validator,
	requireAuthz handler.RequireAuthz,
	tx db.TxContext,
	f AddLoginIDInteractionFlow,
) http.Handler {
	h := &AddLoginIDHandler{
		Validator:    v,
		TxContext:    tx,
		Interactions: f,
	}
	return requireAuthz(h, h)
}

func newAddLoginIDHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		provideAddLoginIDHandler,
		wire.Bind(new(AddLoginIDInteractionFlow), new(*interactionflows.AuthAPIFlow)),
	)
	return nil
}
