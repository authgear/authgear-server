package identity

import (
	"net/url"
	"strconv"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/web3"
)

type SIWE struct {
	ID        string     `json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	UserID    string     `json:"user_id"`
	ChainID   int        `json:"chain_id"`
	Address   web3.EIP55 `json:"address"`

	Data *model.SIWEVerifiedData `json:"data"`
}

func (i *SIWE) ToInfo() *Info {
	return &Info{
		ID:        i.ID,
		UserID:    i.UserID,
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
		Type:      model.IdentityTypeSIWE,

		SIWE: i,
	}
}

func (i *SIWE) ToERC681() (*web3.EIP681, error) {
	return web3.NewEIP681(i.ChainID, i.Address.String(), url.Values{})
}

func (i *SIWE) ToContractID() (*web3.ContractID, error) {
	return web3.NewContractID("ethereum", strconv.Itoa(i.ChainID), i.Address.String(), url.Values{})
}
