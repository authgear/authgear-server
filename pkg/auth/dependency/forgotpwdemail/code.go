package forgotpwdemail

import (
	"crypto/sha256"
	"fmt"
	"io"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
)

type CodeGenerator struct {
	MasterKey string
}

func (c *CodeGenerator) Generate(
	authInfo authinfo.AuthInfo,
	email string,
	hashedPassword []byte,
	expireAt time.Time,
) string {
	h := sha256.New()
	io.WriteString(h, c.MasterKey)
	io.WriteString(h, authInfo.ID)
	if email != "" {
		io.WriteString(h, email)
	}
	io.WriteString(h, expireAt.Format(time.RFC3339))
	if len(hashedPassword) > 0 {
		h.Write(hashedPassword)
	}
	if authInfo.LastLoginAt != nil && !authInfo.LastLoginAt.IsZero() {
		io.WriteString(h, authInfo.LastLoginAt.Format(time.RFC3339))
	}

	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)[0:8]
}
