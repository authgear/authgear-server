package loader

import (
	"github.com/authgear/authgear-server/pkg/admin/model"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type VerificationService interface {
	GetClaims(userID string) ([]*verification.Claim, error)
}

type VerificationLoader struct {
	Verification VerificationService
}

func (l *VerificationLoader) Get(userID string) *graphqlutil.Lazy {
	return graphqlutil.NewLazy(func() (interface{}, error) {
		claims, err := l.Verification.GetClaims(userID)
		if err != nil {
			return nil, err
		}

		var models []model.Claim
		for _, c := range claims {
			models = append(models, model.Claim{
				Name:  c.Name,
				Value: c.Value,
			})
		}
		return models, nil
	})
}
