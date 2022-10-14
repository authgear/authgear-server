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

func (t *ContractID) Clone() *ContractID {
	cloned := *t
	if t.Query != nil {
		cloned.Query = make(url.Values)
		for key, val := range t.Query {
			slice := make([]string, len(val))
			copy(slice, val)
			cloned.Query[key] = slice
		}
	}
	return &cloned
}

func (t *ContractID) StripQuery() *ContractID {
	cloned := *t
	cloned.Query = nil
	return &cloned
}

func (t *ContractID) URL() (*url.URL, error) {
	switch t.Blockchain {
	case "ethereum":

		chainID, err := strconv.Atoi(t.Network)
		if err != nil {
			return nil, err
		}

		eip681, err := NewEIP681(chainID, t.Address.String(), t.Query)
		if err != nil {
			return nil, err
		}

		return eip681.URL(), nil
	default:
		return nil, fmt.Errorf("contract_id: unsupported blockchain: %s", t.Blockchain)
	}
}

func (t *ContractID) String() string {
	u, err := t.URL()
	if err != nil {
		panic(err)
	}
	return u.String()
}

func (t ContractID) MarshalText() ([]byte, error) {
	u, err := t.URL()
	if err != nil {
		return nil, err
	}
	return []byte(u.String()), nil
}

func (t *ContractID) UnmarshalText(text []byte) error {
	parsed, err := ParseContractID(string(text))
	if err != nil {
		return err
	}
	*t = *parsed
	return nil
}
