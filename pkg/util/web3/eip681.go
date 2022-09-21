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
	ChainID int
	Address string
}

func (eip681 *EIP681) URL() *url.URL {
	return &url.URL{
		Scheme: "ethereum",
		Opaque: fmt.Sprintf("%s@%d", eip681.Address, eip681.ChainID),
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

	addressURI := strings.Split(protocolURI[1], "@")

	if len(addressURI) != 2 {
		return nil, fmt.Errorf("invalid uri: %s", uri)
	}

	addressString := addressURI[0]
	address, err := hexstring.Parse(addressString)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %s", address)
	}

	chainID, err := strconv.Atoi(addressURI[1])
	if err != nil {
		return nil, err
	}

	if chainID < 0 {
		return nil, fmt.Errorf("chain id cannot be negative: %d", chainID)
	}

	return &EIP681{
		ChainID: chainID,
		Address: string(address),
	}, nil

}
