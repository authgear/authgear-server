package appresource

import "github.com/authgear/authgear-server/pkg/api/apierrors"


var ResouceTooLarge = apierrors.RequestEntityTooLarge.WithReason("ResourceTooLarge")
