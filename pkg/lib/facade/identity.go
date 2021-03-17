package facade

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

type IdentityFacade struct {
	Coordinator *Coordinator
}

func (i IdentityFacade) Get(userID string, typ authn.IdentityType, id string) (*identity.Info, error) {
	return i.Coordinator.IdentityGet(userID, typ, id)
}

func (i IdentityFacade) GetBySpec(spec *identity.Spec) (*identity.Info, error) {
	return i.Coordinator.IdentityGetBySpec(spec)
}

func (i IdentityFacade) ListByUser(userID string) ([]*identity.Info, error) {
	return i.Coordinator.IdentityListByUser(userID)
}

func (i IdentityFacade) ListByClaim(name string, value string) ([]*identity.Info, error) {
	return i.Coordinator.IdentityListByClaim(name, value)
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

func (i IdentityFacade) Update(info *identity.Info) error {
	return i.Coordinator.IdentityUpdate(info)
}

func (i IdentityFacade) Delete(is *identity.Info) error {
	return i.Coordinator.IdentityDelete(is)
}

func (i IdentityFacade) CheckDuplicated(info *identity.Info) (*identity.Info, error) {
	return i.Coordinator.IdentityCheckDuplicated(info)
}
