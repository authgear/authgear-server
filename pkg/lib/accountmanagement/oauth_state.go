package accountmanagement

import (
	"net/url"
	"strings"

	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

func GenerateRandomState() string {
	// Some provider has a hard-limit on the length of the state.
	// Here we use 32 which is observed to be short enough.
	state := rand.StringWithAlphabet(32, base32.Alphabet, rand.SecureRand)
	return state
}

func ExtractStateFromQuery(query string) (state string, err error) {
	// query may start with a ?, remove it.
	query = strings.TrimPrefix(query, "?")
	form, err := url.ParseQuery(query)
	if err != nil {
		return
	}

	state = form.Get("state")
	return
}
