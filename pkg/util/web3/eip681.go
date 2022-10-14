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

func (t *EIP681) URL() *url.URL {
	return &url.URL{
		Scheme:   "ethereum",
		Opaque:   fmt.Sprintf("%s@%d", t.Address, t.ChainID),
		RawQuery: t.Query.Encode(),
	}
}

func (t *EIP681) String() string {
	return t.URL().String()
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

func (t EIP681) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *EIP681) UnmarshalText(text []byte) error {
	parsed, err := ParseEIP681(string(text))
	if err != nil {
		return err
	}
	*t = *parsed
	return nil
}
