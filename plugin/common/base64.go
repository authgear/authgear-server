package common

import (
	"bytes"
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

	buf := bytes.Buffer{}
	encoder := base64.NewEncoder(base64.URLEncoding, &buf)
	encoder.Write(out)
	return buf.String(), nil
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
