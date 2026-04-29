package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"
	portalmodel "github.com/authgear/authgear-server/pkg/portal/model"
	portalservice "github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/portal/session"
)

type fakeCollaboratorDatabase struct {
	inTx atomic.Bool
}

func (f *fakeCollaboratorDatabase) WithTx(ctx context.Context, do func(context.Context) error) error {
	if !f.inTx.CompareAndSwap(false, true) {
		return errors.New("nested fake transaction")
	}
	defer f.inTx.Store(false)
	return do(ctx)
}

type fakeCollaboratorStore struct {
	listed            []*portalmodel.Collaborator
	existingByID      map[string]*portalmodel.Collaborator
	existingByAppUser map[string]*portalmodel.Collaborator
	created           []*portalmodel.Collaborator
	deleted           []*portalmodel.Collaborator
	updated           []*portalmodel.Collaborator
	now               time.Time
}

func (f *fakeCollaboratorStore) ListCollaborators(_ context.Context, appID string) ([]*portalmodel.Collaborator, error) {
	var out []*portalmodel.Collaborator
	for _, c := range f.listed {
		if c.AppID == appID {
			out = append(out, c)
		}
	}
	return out, nil
}

func (f *fakeCollaboratorStore) GetCollaborator(_ context.Context, id string) (*portalmodel.Collaborator, error) {
	if c, ok := f.existingByID[id]; ok {
		return c, nil
	}
	return nil, portalservice.ErrCollaboratorNotFound
}

func (f *fakeCollaboratorStore) GetCollaboratorByAppAndUser(_ context.Context, appID string, userID string) (*portalmodel.Collaborator, error) {
	if c, ok := f.existingByAppUser[appID+":"+userID]; ok {
		return c, nil
	}
	return nil, portalservice.ErrCollaboratorNotFound
}

func (f *fakeCollaboratorStore) NewCollaborator(appID string, userID string, role portalmodel.CollaboratorRole) *portalmodel.Collaborator {
	return &portalmodel.Collaborator{
		ID:        "new-collaborator",
		AppID:     appID,
		UserID:    userID,
		CreatedAt: f.now,
		Role:      role,
	}
}

func (f *fakeCollaboratorStore) CreateCollaborator(_ context.Context, c *portalmodel.Collaborator) error {
	f.created = append(f.created, c)
	return nil
}

func (f *fakeCollaboratorStore) DeleteCollaborator(_ context.Context, c *portalmodel.Collaborator) error {
	f.deleted = append(f.deleted, c)
	return nil
}

func (f *fakeCollaboratorStore) UpdateCollaborator(_ context.Context, c *portalmodel.Collaborator) error {
	f.updated = append(f.updated, c)
	if existing, ok := f.existingByID[c.ID]; ok {
		existing.Role = c.Role
	}
	return nil
}

type fakeSiteadminAdminAPI struct {
	serverURL string
}

func (f *fakeSiteadminAdminAPI) SelfDirector(_ context.Context, _ string, _ portalservice.Usage) (func(*http.Request), error) {
	return func(r *http.Request) {
		u, _ := url.Parse(f.serverURL)
		r.URL = u
		r.Host = u.Host
	}, nil
}

func ctxWithCollaboratorSession() context.Context {
	return session.WithSessionInfo(context.Background(), &model.SessionInfo{
		IsValid: true,
		UserID:  "actor-user",
	})
}

func TestCollaboratorService(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	makeAdminAPI := func(svr *httptest.Server) *AdminAPIService {
		return &AdminAPIService{
			AdminAPI:   &fakeSiteadminAdminAPI{serverURL: svr.URL},
			HTTPClient: SiteAdminHTTPClient{Client: &http.Client{}},
		}
	}

	Convey("CollaboratorService", t, func() {
		Convey("ListCollaborators resolves emails after the DB transaction finishes", func() {
			db := &fakeCollaboratorDatabase{}
			store := &fakeCollaboratorStore{
				listed: []*portalmodel.Collaborator{
					{ID: "collab-1", AppID: "app-1", UserID: "user-1", CreatedAt: now, Role: portalmodel.CollaboratorRoleOwner},
				},
			}
			globalID := relay.ToGlobalID("User", "user-1")
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if db.inTx.Load() {
					http.Error(w, "admin api called while transaction is open", http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(getNodesResponse(
					map[string]interface{}{"id": globalID, "standardAttributes": map[string]interface{}{"email": "alice@example.com"}},
				))
			}))
			defer svr.Close()

			svc := &CollaboratorService{
				GlobalDatabase: db,
				Store:          store,
				AdminAPI:       makeAdminAPI(svr),
			}

			result, err := svc.ListCollaborators(ctxWithCollaboratorSession(), "app-1")

			So(err, ShouldBeNil)
			So(result, ShouldResemble, []siteadmin.Collaborator{{
				Id:        "collab-1",
				AppId:     "app-1",
				UserId:    "user-1",
				UserEmail: "alice@example.com",
				Role:      siteadmin.Owner,
				CreatedAt: now,
			}})
		})

		Convey("AddCollaborator looks up by email before opening a DB transaction", func() {
			db := &fakeCollaboratorDatabase{}
			store := &fakeCollaboratorStore{
				existingByAppUser: map[string]*portalmodel.Collaborator{},
				now:               now,
			}
			globalID := relay.ToGlobalID("User", "user-2")
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if db.inTx.Load() {
					http.Error(w, "admin api called while transaction is open", http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(getUsersByEmailResponse(globalID))
			}))
			defer svr.Close()

			svc := &CollaboratorService{
				GlobalDatabase: db,
				Store:          store,
				AdminAPI:       makeAdminAPI(svr),
			}

			collaborator, err := svc.AddCollaborator(ctxWithCollaboratorSession(), "app-1", "bob@example.com")

			So(err, ShouldBeNil)
			So(collaborator, ShouldResemble, &siteadmin.Collaborator{
				Id:        "new-collaborator",
				AppId:     "app-1",
				UserId:    "user-2",
				UserEmail: "bob@example.com",
				Role:      siteadmin.Editor,
				CreatedAt: now,
			})
			So(store.created, ShouldHaveLength, 1)
		})

		Convey("RemoveCollaborator rejects cross-app deletes as not found", func() {
			db := &fakeCollaboratorDatabase{}
			store := &fakeCollaboratorStore{
				existingByID: map[string]*portalmodel.Collaborator{
					"collab-1": {ID: "collab-1", AppID: "other-app", UserID: "user-1", CreatedAt: now, Role: portalmodel.CollaboratorRoleEditor},
				},
			}
			svc := &CollaboratorService{
				GlobalDatabase: db,
				Store:          store,
				AdminAPI:       nil,
			}

			err := svc.RemoveCollaborator(ctxWithCollaboratorSession(), "app-1", "collab-1")

			So(err, ShouldEqual, portalservice.ErrCollaboratorNotFound)
			So(store.deleted, ShouldBeEmpty)
		})

		Convey("RemoveCollaborator rejects owner deletion", func() {
			db := &fakeCollaboratorDatabase{}
			store := &fakeCollaboratorStore{
				existingByID: map[string]*portalmodel.Collaborator{
					"collab-1": {ID: "collab-1", AppID: "app-1", UserID: "user-1", CreatedAt: now, Role: portalmodel.CollaboratorRoleOwner},
				},
			}
			svc := &CollaboratorService{
				GlobalDatabase: db,
				Store:          store,
			}

			err := svc.RemoveCollaborator(ctxWithCollaboratorSession(), "app-1", "collab-1")

			So(err, ShouldEqual, ErrCollaboratorOwnerDeletion)
			So(store.deleted, ShouldBeEmpty)
		})

		Convey("PromoteCollaborator succeeds and resolves email outside TX", func() {
			db := &fakeCollaboratorDatabase{}
			store := &fakeCollaboratorStore{
				listed: []*portalmodel.Collaborator{
					{ID: "owner-1", AppID: "app-1", UserID: "user-owner", CreatedAt: now, Role: portalmodel.CollaboratorRoleOwner},
					{ID: "editor-1", AppID: "app-1", UserID: "user-editor", CreatedAt: now, Role: portalmodel.CollaboratorRoleEditor},
				},
				existingByID: map[string]*portalmodel.Collaborator{
					"editor-1": {ID: "editor-1", AppID: "app-1", UserID: "user-editor", CreatedAt: now, Role: portalmodel.CollaboratorRoleEditor},
				},
			}
			globalID := relay.ToGlobalID("User", "user-editor")
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if db.inTx.Load() {
					http.Error(w, "admin api called while transaction is open", http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(getNodesResponse(
					map[string]interface{}{"id": globalID, "standardAttributes": map[string]interface{}{"email": "editor@example.com"}},
				))
			}))
			defer svr.Close()

			svc := &CollaboratorService{
				GlobalDatabase: db,
				Store:          store,
				AdminAPI:       makeAdminAPI(svr),
			}

			result, err := svc.PromoteCollaborator(ctxWithCollaboratorSession(), "app-1", "editor-1")

			So(err, ShouldBeNil)
			So(result, ShouldResemble, &siteadmin.Collaborator{
				Id:        "editor-1",
				AppId:     "app-1",
				UserId:    "user-editor",
				UserEmail: "editor@example.com",
				Role:      siteadmin.Owner,
				CreatedAt: now,
			})
			So(store.updated, ShouldHaveLength, 2)
			So(store.updated[0].ID, ShouldEqual, "editor-1")
			So(store.updated[0].Role, ShouldEqual, portalmodel.CollaboratorRoleOwner)
			So(store.updated[1].ID, ShouldEqual, "owner-1")
			So(store.updated[1].Role, ShouldEqual, portalmodel.CollaboratorRoleEditor)
		})

		Convey("PromoteCollaborator returns not found when collaboratorID is missing", func() {
			db := &fakeCollaboratorDatabase{}
			store := &fakeCollaboratorStore{
				existingByID: map[string]*portalmodel.Collaborator{},
			}
			svc := &CollaboratorService{
				GlobalDatabase: db,
				Store:          store,
			}

			_, err := svc.PromoteCollaborator(ctxWithCollaboratorSession(), "app-1", "no-such-id")

			So(err, ShouldEqual, portalservice.ErrCollaboratorNotFound)
			So(store.updated, ShouldBeEmpty)
		})

		Convey("PromoteCollaborator returns not found on cross-app access", func() {
			db := &fakeCollaboratorDatabase{}
			store := &fakeCollaboratorStore{
				existingByID: map[string]*portalmodel.Collaborator{
					"collab-1": {ID: "collab-1", AppID: "other-app", UserID: "user-1", CreatedAt: now, Role: portalmodel.CollaboratorRoleEditor},
				},
			}
			svc := &CollaboratorService{
				GlobalDatabase: db,
				Store:          store,
			}

			_, err := svc.PromoteCollaborator(ctxWithCollaboratorSession(), "app-1", "collab-1")

			So(err, ShouldEqual, portalservice.ErrCollaboratorNotFound)
			So(store.updated, ShouldBeEmpty)
		})

		Convey("PromoteCollaborator returns AlreadyOwner when target is already owner", func() {
			db := &fakeCollaboratorDatabase{}
			store := &fakeCollaboratorStore{
				existingByID: map[string]*portalmodel.Collaborator{
					"collab-1": {ID: "collab-1", AppID: "app-1", UserID: "user-1", CreatedAt: now, Role: portalmodel.CollaboratorRoleOwner},
				},
			}
			svc := &CollaboratorService{
				GlobalDatabase: db,
				Store:          store,
			}

			_, err := svc.PromoteCollaborator(ctxWithCollaboratorSession(), "app-1", "collab-1")

			So(err, ShouldEqual, ErrCollaboratorAlreadyOwner)
			So(store.updated, ShouldBeEmpty)
		})
	})
}
