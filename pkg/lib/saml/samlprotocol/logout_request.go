package samlprotocol

import (
	"bytes"
	"encoding/xml"

	xrv "github.com/mattermost/xml-roundtrip-validator"
)

func ParseLogoutRequest(input []byte) (*LogoutRequest, error) {
	var req LogoutRequest
	if err := xrv.Validate(bytes.NewReader(input)); err != nil {
		return nil, err
	}

	if err := xml.Unmarshal(input, &req); err != nil {
		return nil, err
	}

	return &req, nil
}
