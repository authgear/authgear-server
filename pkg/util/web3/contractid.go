package web3

import (
	"fmt"
	"net/url"
	"strconv"
)

type ContractID struct {
	Blockchain string
	Network    string
	Address    EIP55
	Query      url.Values
}

func NewContractID(blockchain string, network string, address string, query url.Values) (*ContractID, error) {
	hexaddr, err := NewEIP55(address)
	if err != nil {
		return nil, err
	}

	return &ContractID{
		Blockchain: blockchain,
		Network:    network,
		Address:    hexaddr,
		Query:      query,
	}, nil
}

func ParseContractID(contractURL string) (*ContractID, error) {
	curl, err := url.Parse(contractURL)
	if err != nil {
		return nil, err
	}

	protocol := curl.Scheme

	switch protocol {
	case "ethereum":
		eip681, err := ParseEIP681(contractURL)
		if err != nil {
			return nil, err
		}

		return NewContractID("ethereum", strconv.Itoa(eip681.ChainID), eip681.Address.String(), eip681.Query)
	default:
		return nil, fmt.Errorf("contract_id: unknown protocol: %s", protocol)
	}
}

func (cid *ContractID) URL() (*url.URL, error) {
	switch cid.Blockchain {
	case "ethereum":

		chainID, err := strconv.Atoi(cid.Network)
		if err != nil {
			return nil, err
		}

		eip681, err := NewEIP681(chainID, cid.Address.String(), cid.Query)
		if err != nil {
			return nil, err
		}

		return eip681.URL(), nil
	default:
		return nil, fmt.Errorf("contract_id: unsupported blockchain: %s", cid.Blockchain)
	}
}
