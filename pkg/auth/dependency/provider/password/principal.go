package password

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/skygeario/skygear-server/pkg/server/uuid"
)

type Principal struct {
	ID             string
	UserID         string
	AuthData       interface{}
	PlainPassword  string
	HashedPassword []byte
}

func NewPrincipal() Principal {
	return Principal{
		ID: uuid.New(),
	}
}

// NewUniqueAuthData converts authData by a list of authData depending on authRecordKeys
// example 1: authRecordKeys = username, email
// - if authData is { "username": "john.doe" }, output is [{ "username": "john.doe" }]
// - if authData is { "username": "john.doe", "email": "john.doe@example.com" }, output is [{ "username": "john.doe" }, { "email": "john.doe@example.com" }]
// example 2: authRecordKeys = (username, nickname), email
// - if authData is { "username": "john.doe", "email": "john.doe@example.com", "nickname": "john.doe" },
// output is [{ "username": "john.doe", "nickname": "john.doe" }, { "email": "john.doe@example.com" }]
func NewUniqueAuthData(authRecordKeys [][]string, authData map[string]interface{}) []map[string]interface{} {
	outputs := make([]map[string]interface{}, 0)

	for _, ks := range authRecordKeys {
		m := make(map[string]interface{})
		for _, k := range ks {
			for dk := range authData {
				if k == dk && authData[dk] != nil {
					m[k] = authData[dk]
				}
			}
		}
		if len(m) != 0 { // avoid empty map
			outputs = append(outputs, m)
		}
	}

	return outputs
}

func (p Principal) IsSamePassword(password string) bool {
	return bcrypt.CompareHashAndPassword(p.HashedPassword, []byte(password)) == nil
}
