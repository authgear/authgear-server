package samlprotocol

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/util/rand"
)

const (
	idAlphabet string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// It must start with a letter or underscore, and can only contain letters, digits, underscores, hyphens, and periods.
// https://www.w3.org/TR/2012/REC-xmlschema11-2-20120405/datatypes.html#ID
// https://www.w3.org/TR/2012/REC-xmlschema11-2-20120405/datatypes.html#NCName

func generateID(prefix string) string {
	id := rand.StringWithAlphabet(32, idAlphabet, rand.SecureRand)
	return fmt.Sprintf("%s_%s", prefix, id)
}

func GenerateResponseID() string {
	return generateID("samlresponse")
}
