package pgsearch

import (
	"context"

	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/searchdb"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type Service struct {
	AppID    *config.AppID
	Store    *Store
	Database *searchdb.Handle
}

func (s *Service) QueryUser(
	ctx context.Context,
	searchKeyword string,
	filters user.FilterOptions,
	sortOption user.SortOption,
	pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, error) {
	var refs []apimodel.PageItemRef
	err := s.withReadOnlyTx(ctx, func(ctx context.Context) error {
		var err error
		refs, err = s.Store.QueryUser(ctx, searchKeyword, filters, sortOption, pageArgs)
		return err
	})
	if err != nil {
		return nil, err
	}

	return refs, nil
}

func (s *Service) ReindexUser(
	ctx context.Context, user *apimodel.SearchUserSource) error {
	err := s.withTx(ctx, func(ctx context.Context) error {
		return s.Store.UpsertUsers(ctx, []*apimodel.SearchUserSource{user})
	})
	return err
}

func (s *Service) DeleteUser(
	ctx context.Context, userID string) error {
	err := s.withTx(ctx, func(ctx context.Context) error {
		return s.Store.DeleteUser(ctx, string(*s.AppID), userID)
	})
	return err
}

func (s *Service) withTx(ctx context.Context, do func(ctx context.Context) error) error {
	if s.Database == nil {
		return ErrMissingCredential
	}
	return s.Database.WithTx(ctx, func(ctx context.Context) error {
		return do(ctx)
	})
}

func (s *Service) withReadOnlyTx(ctx context.Context, do func(ctx context.Context) error) error {
	if s.Database == nil {
		return ErrMissingCredential
	}
	return s.Database.ReadOnly(ctx, func(ctx context.Context) error {
		return do(ctx)
	})
}
