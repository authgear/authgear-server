package password

import "github.com/authgear/authgear-server/pkg/api/apierrors"

var PasswordPolicyViolated apierrors.Kind = apierrors.Invalid.WithReason("PasswordPolicyViolated")
var PasswordExpiryForceChange apierrors.Kind = apierrors.Invalid.WithReason("PasswordExpiryForceChange")
