package webapp

import (
	"mime"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/secretcode"
)

// This file contains utility functions used by v2 code.
// Ideally, we should move these back to v2.
// So please do not add new things in this file.

func FormatRecoveryCodes(recoveryCodes []string) []string {
	out := make([]string, len(recoveryCodes))
	for i, code := range recoveryCodes {
		out[i] = secretcode.RecoveryCode.FormatForHuman(code)
	}
	return out
}

func SetRecoveryCodeAttachmentHeaders(w http.ResponseWriter) {
	// No need to use FormatMediaType because the value is constant.
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
		"filename": "recovery-codes.txt",
	}))
}
