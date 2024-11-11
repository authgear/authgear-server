package facade

import (
	"context"

	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

type IdentityFacade struct {
	Coordinator *Coordinator
}

func (i IdentityFacade) Get(ctx context.Context, id string) (*identity.Info, error) {
	return i.Coordinator.IdentityGet(ctx, id)
}

func (i IdentityFacade) SearchBySpec(ctx context.Context, spec *identity.Spec) (exactMatch *identity.Info, otherMatches []*identity.Info, err error) {
	return i.Coordinator.IdentitySearchBySpec(ctx, spec)
}

func (i IdentityFacade) ListByUser(ctx context.Context, userID string) ([]*identity.Info, error) {
	return i.Coordinator.IdentityListByUser(ctx, userID)
}

func (i IdentityFacade) ListIdentitiesThatHaveStandardAttributes(ctx context.Context, userID string) ([]*identity.Info, error) {
	return i.Coordinator.IdentityListIdentitiesThatHaveStandardAttributes(ctx, userID)
}

func (i IdentityFacade) ListByClaim(ctx context.Context, name string, value string) ([]*identity.Info, error) {
	return i.Coordinator.IdentityListByClaim(ctx, name, value)
}

func (i IdentityFacade) ListRefsByUsers(ctx context.Context, userIDs []string, identityType *apimodel.IdentityType) ([]*apimodel.IdentityRef, error) {
	return i.Coordinator.IdentityListRefsByUsers(ctx, userIDs, identityType)
}

func (i IdentityFacade) New(ctx context.Context, userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error) {
	return i.Coordinator.IdentityNew(ctx, userID, spec, options)
}

func (i IdentityFacade) UpdateWithSpec(ctx context.Context, is *identity.Info, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error) {
	return i.Coordinator.IdentityUpdateWithSpec(ctx, is, spec, options)
}

func (i IdentityFacade) Create(ctx context.Context, is *identity.Info) error {
	return i.Coordinator.IdentityCreate(ctx, is)
}

func (i IdentityFacade) CreateByAdmin(ctx context.Context, userID string, spec *identity.Spec, password string) (*identity.Info, error) {
	return i.Coordinator.IdentityCreateByAdmin(ctx, userID, spec, password)
}

func (i IdentityFacade) Update(ctx context.Context, oldInfo *identity.Info, newInfo *identity.Info) error {
	return i.Coordinator.IdentityUpdate(ctx, oldInfo, newInfo)
}

func (i IdentityFacade) Delete(ctx context.Context, is *identity.Info) error {
	return i.Coordinator.IdentityDelete(ctx, is)
}

func (i IdentityFacade) CheckDuplicated(ctx context.Context, info *identity.Info) (*identity.Info, error) {
	return i.Coordinator.IdentityCheckDuplicated(ctx, info)
}

func (i IdentityFacade) CheckDuplicatedByUniqueKey(ctx context.Context, info *identity.Info) (*identity.Info, error) {
	return i.Coordinator.IdentityCheckDuplicatedByUniqueKey(ctx, info)
}
