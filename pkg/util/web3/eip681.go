package web3

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// https://eips.ethereum.org/EIPS/eip-681
type EIP681 struct {
	ChainID int
	Address EIP55
	Query   url.Values
}

func (eip681 *EIP681) URL() *url.URL {
	return &url.URL{
		Scheme:   "ethereum",
		Opaque:   fmt.Sprintf("%s@%d", eip681.Address, eip681.ChainID),
		RawQuery: eip681.Query.Encode(),
	}
}

func NewEIP681(chainID int, address string, query url.Values) (*EIP681, error) {

	if chainID <= 0 {
		return nil, fmt.Errorf("chain ID must be positive")
	}

	addrHex, err := NewEIP55(address)
	if err != nil {
		return nil, err
	}

	return &EIP681{
		ChainID: chainID,
		Address: addrHex,
		Query:   query,
	}, nil
}

func ParseEIP681(uri string) (*EIP681, error) {

	url, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid uri: %s", uri)
	}

	if url.Scheme != "ethereum" {
		return nil, fmt.Errorf("invalid protocol: %s", url.Scheme)
	}

	addressURI := strings.Split(url.Opaque, "@")

	if len(addressURI) != 2 {
		return nil, fmt.Errorf("invalid uri: %s", uri)
	}

	addressString := addressURI[0]
	address, err := NewEIP55(addressString)
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
		Address: address,
		Query:   url.Query(),
	}, nil

}
