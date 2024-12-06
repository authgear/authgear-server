package webapp

import (
	"github.com/authgear/authgear-server/pkg/util/template"
)

// This file contains the v1 templates that is used by v2 code.
// Ideally we should move the templates back to v2 package.
// So please do not add new things here.

var TemplateWebDownloadRecoveryCodeTXT = template.RegisterPlainText(
	"web/download_recovery_code.txt",
	plainTextComponents...,
)
