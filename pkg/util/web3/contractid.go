package web3

import (
	"fmt"
	"net/url"
	"strconv"
)

type ContractID struct {
	Blockchain      string
	Network         string
	ContractAddress string
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

		return &ContractID{
			Blockchain:      "ethereum",
			Network:         strconv.Itoa(eip681.ChainID),
			ContractAddress: eip681.Address,
		}, nil
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

		if chainID <= 0 {
			return nil, err
		}

		eip681 := &EIP681{
			ChainID: chainID,
			Address: cid.ContractAddress,
		}

		return eip681.URL(), nil
	default:
		return nil, fmt.Errorf("contract_id: unsupported blockchain: %s", cid.Blockchain)
	}
}
