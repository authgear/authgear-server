package event

import (
	"context"
	"errors"
	"reflect"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

type ResolverUserQueries interface {
	Get(ctx context.Context, id string, role accesscontrol.Role) (*model.User, error)
}

type ResolverImpl struct {
	Users ResolverUserQueries
}

func (r *ResolverImpl) Resolve(ctx context.Context, anything interface{}) (err error) {
	struc := reflect.ValueOf(anything).Elem()
	typ := struc.Type()

	fields := reflect.VisibleFields(typ)
	for i, refField := range fields {
		if jsonName, ok := refField.Tag.Lookup("resolve"); ok {
			for j, targetField := range fields {
				if name, ok := targetField.Tag.Lookup("json"); ok {
					if jsonName == name {
						userRef := struc.Field(i).Interface().(model.UserRef)

						var u *model.User
						u, err = r.Users.Get(ctx, userRef.ID, accesscontrol.RoleGreatest)
						if errors.Is(err, user.ErrUserNotFound) {
							continue
						}

						if err != nil {
							return
						}

						struc.Field(j).Set(reflect.ValueOf(*u))
					}
				}
			}
		}
	}

	return
}
