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
	Query   *url.Values
}

func (eip681 *EIP681) URL() *url.URL {
	query := ""
	if eip681.Query != nil {
		query = eip681.Query.Encode()
	}
	return &url.URL{
		Scheme:   "ethereum",
		Opaque:   fmt.Sprintf("%s@%d", eip681.Address, eip681.ChainID),
		RawQuery: query,
	}
}

func NewEIP681(chainID int, address string, query *url.Values) (*EIP681, error) {

	if chainID <= 0 {
		return nil, fmt.Errorf("chain ID must be positive")
	}

	addrHex, err := hexstring.Parse(address)
	if err != nil {
		return nil, err
	}

	return &EIP681{
		ChainID: chainID,
		Address: addrHex.String(),
		Query:   query,
	}, nil
}

func ParseEIP681(uri string) (*EIP681, error) {
	queryString := strings.Split(uri, "?")
	protocolURI := strings.Split(queryString[0], ":")

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

	values := (*url.Values)(nil)
	if len(queryString) > 1 {
		query, err := url.ParseQuery(queryString[1])
		if err != nil {
			return nil, fmt.Errorf("invalid query: %s", queryString[1])
		}
		values = &query
	}

	return &EIP681{
		ChainID: chainID,
		Address: string(address),
		Query:   values,
	}, nil

}
