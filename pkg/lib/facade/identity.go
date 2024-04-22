package facade

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
)

type IdentityFacade struct {
	Coordinator *Coordinator
}

func (i IdentityFacade) Get(id string) (*identity.Info, error) {
	return i.Coordinator.IdentityGet(id)
}

func (i IdentityFacade) SearchBySpec(spec *identity.Spec) (exactMatch *identity.Info, otherMatches []*identity.Info, err error) {
	return i.Coordinator.IdentitySearchBySpec(spec)
}

func (i IdentityFacade) ListByUser(userID string) ([]*identity.Info, error) {
	return i.Coordinator.IdentityListByUser(userID)
}

func (i IdentityFacade) ListByClaim(name string, value string) ([]*identity.Info, error) {
	return i.Coordinator.IdentityListByClaim(name, value)
}

func (i IdentityFacade) ListByClaimJSONPointer(pointer jsonpointer.T, value string) ([]*identity.Info, error) {
	return i.Coordinator.IdentityListByClaimJSONPointer(pointer, value)
}

func (i IdentityFacade) New(userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error) {
	return i.Coordinator.IdentityNew(userID, spec, options)
}

func (i IdentityFacade) UpdateWithSpec(is *identity.Info, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error) {
	return i.Coordinator.IdentityUpdateWithSpec(is, spec, options)
}

func (i IdentityFacade) Create(is *identity.Info) error {
	return i.Coordinator.IdentityCreate(is)
}

func (i IdentityFacade) Update(oldInfo *identity.Info, newInfo *identity.Info) error {
	return i.Coordinator.IdentityUpdate(oldInfo, newInfo)
}

func (i IdentityFacade) Delete(is *identity.Info) error {
	return i.Coordinator.IdentityDelete(is)
}

func (i IdentityFacade) CheckDuplicated(info *identity.Info) (*identity.Info, error) {
	return i.Coordinator.IdentityCheckDuplicated(info)
}
