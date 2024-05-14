package model

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/authgear/authgear-server/pkg/util/web3"
)

type SIWEPublicKey string
type SIWEVerificationRequest struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

type SIWEWallet struct {
	Address web3.EIP55 `json:"address"`
	ChainID int        `json:"chain_id"`
}

type SIWEVerifiedData struct {
	Message          string        `json:"message"`
	Signature        string        `json:"signature"`
	EncodedPublicKey SIWEPublicKey `json:"encoded_public_key"`
}

func NewSIWEPublicKey(k *ecdsa.PublicKey) (SIWEPublicKey, error) {
	if k.Curve != crypto.S256() {
		return "", fmt.Errorf("Invalid curve: %s", k.Curve)
	}
	return SIWEPublicKey(hex.EncodeToString(crypto.CompressPubkey(k))), nil
}

func (k SIWEPublicKey) ECDSA() (*ecdsa.PublicKey, error) {
	hexKey := string(k)

	bytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, err
	}

	key, err := crypto.DecompressPubkey(bytes)
	if err != nil {
		return nil, err
	}

	return key, nil
}
