package web3

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/authgear/authgear-server/pkg/util/hexstring"
)

// https://eips.ethereum.org/EIPS/eip-681
type EIP681 struct {
	ChainID         int
	ContractAddress string
}

func (eip681 *EIP681) URL() *url.URL {
	return &url.URL{
		Scheme: "ethereum",
		Opaque: fmt.Sprintf("%s@%d", eip681.ContractAddress, eip681.ChainID),
	}
}

func ParseEIP681(uri string) (*EIP681, error) {
	protocolURI := strings.Split(uri, ":")

	if len(protocolURI) != 2 {
		return nil, fmt.Errorf("invalid uri: %s", uri)
	}

	if protocolURI[0] != "ethereum" {
		return nil, fmt.Errorf("invalid protocol: %s", protocolURI[0])
	}

	contractURI := strings.Split(protocolURI[1], "@")

	if len(contractURI) != 2 {
		return nil, fmt.Errorf("invalid uri: %s", uri)
	}

	contractAddress := contractURI[0]
	_, err := hexstring.Parse(contractAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid contract address: %s", contractAddress)
	}

	chainID, err := strconv.Atoi(contractURI[1])
	if err != nil {
		return nil, err
	}

	if chainID < 0 {
		return nil, fmt.Errorf("chain id cannot be negative: %d", chainID)
	}

	return &EIP681{
		ChainID:         chainID,
		ContractAddress: contractAddress,
	}, nil

}
