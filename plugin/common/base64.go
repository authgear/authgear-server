package common

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"strings"
)

// EncodeBase64JSON encodes an object into a base64 encoded JSON string
func EncodeBase64JSON(data interface{}) (string, error) {
	out, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	result := base64.StdEncoding.EncodeToString(out)
	return result, nil
}

// DecodeBase64JSON decodes a base64 encoded JSON string into an object
func DecodeBase64JSON(encodedStr string, obj interface{}) error {
	decoder := base64.NewDecoder(base64.URLEncoding, strings.NewReader(encodedStr))
	out, err := ioutil.ReadAll(decoder)
	if err != nil {
		return err
	}

	return json.Unmarshal(out, obj)
}
