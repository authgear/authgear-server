package samlprotocol

import (
	"bytes"
	"encoding/xml"

	xrv "github.com/mattermost/xml-roundtrip-validator"
)

func ParseLogoutResponse(input []byte) (*LogoutResponse, error) {
	var res LogoutResponse
	if err := xrv.Validate(bytes.NewReader(input)); err != nil {
		return nil, err
	}

	if err := xml.Unmarshal(input, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
