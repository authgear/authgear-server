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

func (r *ResolverImpl) Resolve(ctx context.Context, anything any) error {
	return r.resolve(ctx, anything, nil)
}

// ResolveWithUser resolves anything, using u for any resolve:"user" field whose
// UserRef.ID matches u.ID instead of reading the user from the database. Other
// resolve tags, and a resolve:"user" field with a different ID, still read.
func (r *ResolverImpl) ResolveWithUser(ctx context.Context, anything any, u *model.User) error {
	return r.resolve(ctx, anything, u)
}

func (r *ResolverImpl) resolve(ctx context.Context, anything any, override *model.User) (err error) {
	struc := reflect.ValueOf(anything).Elem()
	typ := struc.Type()

	fields := reflect.VisibleFields(typ)
	for i, refField := range fields {
		jsonName, ok := refField.Tag.Lookup("resolve")
		if !ok {
			continue
		}

		for j, targetField := range fields {
			name, ok := targetField.Tag.Lookup("json")
			if !ok || name != jsonName {
				continue
			}

			userRef := struc.Field(i).Interface().(model.UserRef)
			var u *model.User
			u, err = r.resolveUser(ctx, jsonName, userRef, override)
			if errors.Is(err, user.ErrUserNotFound) {
				continue
			}
			if err != nil {
				return
			}

			struc.Field(j).Set(reflect.ValueOf(*u))
		}
	}

	return
}

// resolveUser returns override in place of a database read when override is
// the resolve:"user" field's own user, i.e. the same case ResolveWithUser
// documents. Every other resolve tag, and a resolve:"user" field for a
// different user, still reads from the database.
func (r *ResolverImpl) resolveUser(ctx context.Context, jsonName string, userRef model.UserRef, override *model.User) (*model.User, error) {
	if override != nil && jsonName == "user" && override.ID == userRef.ID {
		return override, nil
	}
	return r.Users.Get(ctx, userRef.ID, accesscontrol.RoleGreatest)
}
