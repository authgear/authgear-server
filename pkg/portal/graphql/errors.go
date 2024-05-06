package graphql

import "github.com/authgear/authgear-server/pkg/api/apierrors"

var Unauthenticated = apierrors.Unauthorized.WithReason("Unauthenticated")
var AccessDenied = apierrors.Forbidden.WithReason("AccessDenied")

var QuotaExceeded = apierrors.Invalid.WithReason("QuotaExceeded")
