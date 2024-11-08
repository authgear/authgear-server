package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	sq "github.com/Masterminds/squirrel"
	relay "github.com/authgear/graphql-go-relay"
	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/resource"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/portal/task/tasks"
	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/rand"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

var ErrCollaboratorNotFound = apierrors.NotFound.WithReason("CollaboratorNotFound").New("collaborator not found")
var ErrCollaboratorSelfDeletion = apierrors.Forbidden.WithReason("CollaboratorSelfDeletion").New("cannot remove self from collaborator")
var ErrCollaboratorDuplicate = apierrors.AlreadyExists.WithReason("CollaboratorDuplicate").New("collaborator duplicate")

var ErrCollaboratorInvitationNotFound = apierrors.NotFound.WithReason("CollaboratorInvitationNotFound").New("collaborator invitation not found")
var ErrCollaboratorInvitationDuplicate = apierrors.AlreadyExists.WithReason("CollaboratorInvitationDuplicate").New("collaborator invitation duplicate")
var ErrCollaboratorInvitationInvalidCode = apierrors.Invalid.WithReason("CollaboratorInvitationInvalidCode").New("collaborator invitation invalid code")

var ErrCollaboratorInvitationInvalidEmail = apierrors.Invalid.WithReason("CollaboratorInvitationInvalidEmail").New("the email with the actor does match the invitee email")

var ErrCollaboratorQuotaExceeded = apierrors.Invalid.WithReason("CollaboratorQuotaExceeded").New("collaborator quota exceeded")

type CollaboratorServiceTaskQueue interface {
	Enqueue(param task.Param)
}

type CollaboratorServiceEndpointsProvider interface {
	AcceptCollaboratorInvitationEndpointURL() *url.URL
}

type CollaboratorServiceAdminAPIService interface {
	SelfDirector(ctx context.Context, actorUserID string, usage Usage) (func(*http.Request), error)
}

type CollaboratorAppConfigService interface {
	ResolveContext(ctx context.Context, appID string) (*config.AppContext, error)
}

type CollaboratorService struct {
	Clock       clock.Clock
	SQLBuilder  *globaldb.SQLBuilder
	SQLExecutor *globaldb.SQLExecutor
	HTTPClient  HTTPClient

	GlobalDatabase *globaldb.Handle

	MailConfig     *portalconfig.MailConfig
	TaskQueue      CollaboratorServiceTaskQueue
	Endpoints      CollaboratorServiceEndpointsProvider
	TemplateEngine *template.Engine
	AdminAPI       CollaboratorServiceAdminAPIService

	AppConfigs CollaboratorAppConfigService
}

func (s *CollaboratorService) selectCollaborator() sq.SelectBuilder {
	return s.SQLBuilder.Select(
		"id",
		"app_id",
		"user_id",
		"created_at",
		"role",
	).From(s.SQLBuilder.TableName("_portal_app_collaborator"))
}

func (s *CollaboratorService) selectCollaboratorInvitation() sq.SelectBuilder {
	return s.SQLBuilder.Select(
		"id",
		"app_id",
		"invited_by",
		"invitee_email",
		"code",
		"created_at",
		"expire_at",
	).From(s.SQLBuilder.TableName("_portal_app_collaborator_invitation"))
}

// ListCollaborators acquires connection.
func (s *CollaboratorService) ListCollaborators(ctx context.Context, appID string) ([]*model.Collaborator, error) {
	q := s.selectCollaborator().Where("app_id = ?", appID)

	var cs []*model.Collaborator
	var err error
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		rows, err := s.SQLExecutor.QueryWith(ctx, q)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			c, err := scanCollaborator(rows)
			if err != nil {
				return err
			}
			cs = append(cs, c)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return cs, nil
}

// ListCollaboratorsByUser acquires connection.
func (s *CollaboratorService) ListCollaboratorsByUser(ctx context.Context, userID string) ([]*model.Collaborator, error) {
	q := s.selectCollaborator().Where("user_id = ?", userID)

	var cs []*model.Collaborator
	var err error
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		rows, err := s.SQLExecutor.QueryWith(ctx, q)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			c, err := scanCollaborator(rows)
			if err != nil {
				return err
			}
			cs = append(cs, c)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return cs, nil
}

// GetProjectOwnerCount acquires connection.
func (s *CollaboratorService) GetProjectOwnerCount(ctx context.Context, userID string) (int, error) {
	q := s.SQLBuilder.Select("count(1)").
		From(s.SQLBuilder.TableName("_portal_app_collaborator")).
		Where("user_id = ?", userID).
		Where("role = ?", string(model.CollaboratorRoleOwner))

	var count int
	var err error
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		row, err := s.SQLExecutor.QueryRowWith(ctx, q)
		if err != nil {
			return err
		}

		err = row.Scan(&count)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetManyProjectOwnerCount acquires connection.
func (s *CollaboratorService) GetManyProjectOwnerCount(ctx context.Context, userIDs []string) ([]int, error) {
	q := s.SQLBuilder.Select(
		"user_id",
		"count(1)",
	).
		From(s.SQLBuilder.TableName("_portal_app_collaborator")).
		Where("user_id = ANY(?)", pq.Array(userIDs)).
		GroupBy("user_id", "role").
		Having("role = ?", string(model.CollaboratorRoleOwner))

	m := make(map[string]int)
	var err error
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		rows, err := s.SQLExecutor.QueryWith(ctx, q)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var userID string
			var count int
			err = rows.Scan(&userID, &count)
			if err != nil {
				return err
			}
			m[userID] = count
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	out := make([]int, len(userIDs))
	for i, userID := range userIDs {
		if count, ok := m[userID]; ok {
			out[i] = count
		} else {
			// By definition, it is zero.
			out[i] = 0
		}
	}

	return out, nil
}

// NewCollaborator does not need connection.
func (s *CollaboratorService) NewCollaborator(appID string, userID string, role model.CollaboratorRole) *model.Collaborator {
	now := s.Clock.NowUTC()
	c := &model.Collaborator{
		ID:        uuid.New(),
		AppID:     appID,
		UserID:    userID,
		CreatedAt: now,
		Role:      role,
	}
	return c
}

// CreateCollaborator assume acquired connection.
func (s *CollaboratorService) CreateCollaborator(ctx context.Context, c *model.Collaborator) error {
	err := s.deleteExpiredInvitations(ctx)
	if err != nil {
		return err
	}

	_, err = s.SQLExecutor.ExecWith(ctx, s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_portal_app_collaborator")).
		Columns(
			"id",
			"app_id",
			"user_id",
			"created_at",
			"role",
		).
		Values(
			c.ID,
			c.AppID,
			c.UserID,
			c.CreatedAt,
			c.Role,
		),
	)
	if err != nil {
		return err
	}

	return nil
}

// GetCollaborator acquires connection.
func (s *CollaboratorService) GetCollaborator(ctx context.Context, id string) (*model.Collaborator, error) {
	q := s.selectCollaborator().Where("id = ?", id)

	var coll *model.Collaborator
	var err error
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		row, err := s.SQLExecutor.QueryRowWith(ctx, q)
		if err != nil {
			return err
		}

		coll, err = scanCollaborator(row)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return coll, nil
}

// GetManyCollaborators acquires connection.
func (s *CollaboratorService) GetManyCollaborators(ctx context.Context, ids []string) ([]*model.Collaborator, error) {
	q := s.selectCollaborator().Where("id = ANY (?)", pq.Array(ids))

	var cs []*model.Collaborator
	var err error
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		rows, err := s.SQLExecutor.QueryWith(ctx, q)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			c, err := scanCollaborator(rows)
			if err != nil {
				return err
			}
			cs = append(cs, c)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return cs, nil
}

// GetCollaboratorByAppAndUser acquires connection.
func (s *CollaboratorService) GetCollaboratorByAppAndUser(ctx context.Context, appID string, userID string) (*model.Collaborator, error) {
	q := s.selectCollaborator().Where("app_id = ? AND user_id = ?", appID, userID)
	var coll *model.Collaborator
	var err error
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		row, err := s.SQLExecutor.QueryRowWith(ctx, q)
		if err != nil {
			return err
		}

		coll, err = scanCollaborator(row)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return coll, nil
}

// DeleteCollaborator acquires connection.
func (s *CollaboratorService) DeleteCollaborator(ctx context.Context, c *model.Collaborator) error {
	sessionInfo := session.GetValidSessionInfo(ctx)
	if c.UserID == sessionInfo.UserID {
		return ErrCollaboratorSelfDeletion
	}

	var err error
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		err = s.deleteExpiredInvitations(ctx)
		if err != nil {
			return err
		}
		_, err = s.SQLExecutor.ExecWith(ctx, s.SQLBuilder.
			Delete(s.SQLBuilder.TableName("_portal_app_collaborator")).
			Where("id = ?", c.ID),
		)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// GetManyInvitations acquires connection.
func (s *CollaboratorService) GetManyInvitations(ctx context.Context, ids []string) ([]*model.CollaboratorInvitation, error) {
	q := s.selectCollaboratorInvitation().Where("id = ANY (?)", pq.Array(ids))

	var is []*model.CollaboratorInvitation
	var err error
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		rows, err := s.SQLExecutor.QueryWith(ctx, q)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			i, err := scanCollaboratorInvitation(rows)
			if err != nil {
				return err
			}
			is = append(is, i)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return is, nil
}

// ListInvitations acquires connection.
func (s *CollaboratorService) ListInvitations(ctx context.Context, appID string) ([]*model.CollaboratorInvitation, error) {
	now := s.Clock.NowUTC()
	q := s.selectCollaboratorInvitation().Where("app_id = ? AND expire_at > ?", appID, now)

	var is []*model.CollaboratorInvitation
	var err error
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		rows, err := s.SQLExecutor.QueryWith(ctx, q)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			i, err := scanCollaboratorInvitation(rows)
			if err != nil {
				return err
			}
			is = append(is, i)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return is, nil
}

// SendInvitation acquires connection.
func (s *CollaboratorService) SendInvitation(
	ctx context.Context,
	appID string,
	inviteeEmail string,
) (*model.CollaboratorInvitation, error) {
	sessionInfo := session.GetValidSessionInfo(ctx)
	invitedBy := sessionInfo.UserID

	err := s.checkQuotaInSend(ctx, appID)
	if err != nil {
		return nil, err
	}

	// TODO(collaborator): Ideally we should prevent sending invitation to existing collaborator.
	// However, this is not harmful to not have it.
	// The collaborator will receive the invitation and they cannot accept it because
	// we have database constraint to enforce this invariant.
	// If Admin API have getUserByClaim, then we can detect this condition here.

	// Check if the invitee has a pending invitation already.
	invitations, err := s.ListInvitations(ctx, appID)
	if err != nil {
		return nil, err
	}
	for _, i := range invitations {
		if i.InviteeEmail == inviteeEmail {
			return nil, ErrCollaboratorInvitationDuplicate
		}
	}

	code := generateCollaboratorInvitationCode()
	now := s.Clock.NowUTC()
	// Expire in 3 days.
	expireAt := now.Add(3 * 24 * time.Hour)

	i := &model.CollaboratorInvitation{
		ID:           uuid.New(),
		AppID:        appID,
		InvitedBy:    invitedBy,
		InviteeEmail: inviteeEmail,
		Code:         code,
		CreatedAt:    now,
		ExpireAt:     expireAt,
	}

	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		err = s.deleteExpiredInvitations(ctx)
		if err != nil {
			return err
		}

		err = s.createCollaboratorInvitation(ctx, i)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	link := s.Endpoints.AcceptCollaboratorInvitationEndpointURL()
	q := link.Query()
	q.Set("code", code)
	link.RawQuery = q.Encode()

	data := map[string]interface{}{
		"AppName": appID,
		"Link":    link,
	}

	textBody, err := s.TemplateEngine.Render(
		resource.TemplateCollaboratorInvitationEmailTXT,
		nil,
		data,
	)
	if err != nil {
		return nil, err
	}

	htmlBody, err := s.TemplateEngine.Render(
		resource.TemplateCollaboratorInvitationEmailHTML,
		nil,
		data,
	)
	if err != nil {
		return nil, err
	}

	s.TaskQueue.Enqueue(&tasks.SendMessagesParam{
		EmailMessages: []mail.SendOptions{
			{
				// TODO(collaborator): We should reuse translation service.
				Sender:    s.MailConfig.Sender,
				ReplyTo:   s.MailConfig.ReplyTo,
				Subject:   "You are invited to collaborate on \"" + appID + "\" in Authgear",
				Recipient: inviteeEmail,
				TextBody:  textBody.String,
				HTMLBody:  htmlBody.String,
			},
		},
	})

	return i, nil
}

// GetInvitation acquires connection.
func (s *CollaboratorService) GetInvitation(ctx context.Context, id string) (*model.CollaboratorInvitation, error) {
	q := s.selectCollaboratorInvitation().Where("id = ?", id)
	var ci *model.CollaboratorInvitation
	var err error
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		row, err := s.SQLExecutor.QueryRowWith(ctx, q)
		if err != nil {
			return err
		}

		ci, err = scanCollaboratorInvitation(row)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return ci, nil
}

// GetInvitationWithCode acquires connection.
func (s *CollaboratorService) GetInvitationWithCode(ctx context.Context, code string) (*model.CollaboratorInvitation, error) {
	now := s.Clock.NowUTC()
	q := s.selectCollaboratorInvitation().Where("code = ? AND expire_at > ?", code, now)

	var is []*model.CollaboratorInvitation
	err := s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		rows, err := s.SQLExecutor.QueryWith(ctx, q)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			i, err := scanCollaboratorInvitation(rows)
			if err != nil {
				return err
			}
			is = append(is, i)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(is) <= 0 {
		return nil, ErrCollaboratorInvitationInvalidCode
	}

	return is[0], nil
}

// DeleteInvitation acquires connection.
func (s *CollaboratorService) DeleteInvitation(ctx context.Context, i *model.CollaboratorInvitation) error {
	var err error
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		err = s.deleteInvitation(ctx, i)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// AcceptInvitation acquires connection.
func (s *CollaboratorService) AcceptInvitation(ctx context.Context, code string) (*model.Collaborator, error) {
	actorID := session.GetValidSessionInfo(ctx).UserID

	invitation, err := s.GetInvitationWithCode(ctx, code)
	if err != nil {
		return nil, err
	}

	err = s.checkQuotaInAccept(ctx, invitation.AppID)
	if err != nil {
		return nil, err
	}

	err = s.CheckInviteeEmail(ctx, invitation, actorID)
	if err != nil {
		return nil, err
	}

	_, err = s.GetCollaboratorByAppAndUser(ctx, invitation.AppID, actorID)
	if err == nil {
		return nil, ErrCollaboratorDuplicate
	}
	if !errors.Is(err, ErrCollaboratorNotFound) {
		return nil, err
	}

	collaborator := s.NewCollaborator(invitation.AppID, actorID, model.CollaboratorRoleEditor)

	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		err = s.deleteInvitation(ctx, invitation)
		if err != nil {
			return err
		}

		err = s.CreateCollaborator(ctx, collaborator)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return collaborator, nil
}

func (s *CollaboratorService) deleteInvitation(ctx context.Context, i *model.CollaboratorInvitation) error {
	var err error
	err = s.deleteExpiredInvitations(ctx)
	if err != nil {
		return err
	}

	_, err = s.SQLExecutor.ExecWith(ctx, s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_portal_app_collaborator_invitation")).
		Where("id = ?", i.ID),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *CollaboratorService) deleteExpiredInvitations(ctx context.Context) error {
	now := s.Clock.NowUTC()
	_, err := s.SQLExecutor.ExecWith(ctx, s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_portal_app_collaborator_invitation")).
		Where("expire_at <= ?", now),
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *CollaboratorService) createCollaboratorInvitation(ctx context.Context, i *model.CollaboratorInvitation) error {
	_, err := s.SQLExecutor.ExecWith(ctx, s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_portal_app_collaborator_invitation")).
		Columns(
			"id",
			"app_id",
			"invited_by",
			"invitee_email",
			"code",
			"created_at",
			"expire_at",
		).
		Values(
			i.ID,
			i.AppID,
			i.InvitedBy,
			i.InviteeEmail,
			i.Code,
			i.CreatedAt,
			i.ExpireAt,
		),
	)
	if err != nil {
		return err
	}

	return nil
}

// CheckInviteeEmail calls HTTP request.
func (s *CollaboratorService) CheckInviteeEmail(ctx context.Context, i *model.CollaboratorInvitation, actorID string) error {
	id := relay.ToGlobalID("User", actorID)

	params := graphqlutil.DoParams{
		OperationName: "getUserNodes",
		Query: `
		query getUserNodes($ids: [ID!]!) {
			nodes(ids: $ids) {
				... on User {
					id
					verifiedClaims {
						name
						value
					}
				}
			}
		}
		`,
		Variables: map[string]interface{}{
			"ids": []interface{}{id},
		},
	}

	r, err := http.NewRequest("POST", "/graphql", nil)
	if err != nil {
		return err
	}

	director, err := s.AdminAPI.SelfDirector(ctx, actorID, UsageInternal)
	if err != nil {
		return err
	}

	director(r)

	result, err := graphqlutil.HTTPDo(s.HTTPClient.Client, r, params)
	if err != nil {
		return err
	}

	if result.HasErrors() {
		return fmt.Errorf("unexpected graphql errors: %v", result.Errors)
	}

	var userModels []*model.User
	data := result.Data.(map[string]interface{})
	nodes := data["nodes"].([]interface{})
	for _, iface := range nodes {
		// It could be null.
		userNode, ok := iface.(map[string]interface{})
		if !ok {
			userModels = append(userModels, nil)
		} else {
			userModel := &model.User{}
			globalID := userNode["id"].(string)
			userModel.ID = globalID

			// Use the last email claim.
			verifiedClaims := userNode["verifiedClaims"].([]interface{})
			for _, iface := range verifiedClaims {
				claim := iface.(map[string]interface{})
				name := claim["name"].(string)
				value := claim["value"].(string)
				if name == "email" {
					userModel.Email = value
				}
			}

			userModels = append(userModels, userModel)
		}
	}

	if len(userModels) != 1 {
		return fmt.Errorf("expected exact one user")
	}

	user := userModels[0]

	if user.Email != i.InviteeEmail {
		return ErrCollaboratorInvitationInvalidEmail
	}

	return nil
}

func (s *CollaboratorService) checkQuotaInSend(ctx context.Context, appID string) error {
	appCtx, err := s.AppConfigs.ResolveContext(ctx, appID)
	if err != nil {
		return err
	}

	collaborators, err := s.ListCollaborators(ctx, appID)
	if err != nil {
		return err
	}

	invitations, err := s.ListInvitations(ctx, appID)
	if err != nil {
		return err
	}

	if appCtx.Config.FeatureConfig.Collaborator.Maximum != nil {
		maximum := *appCtx.Config.FeatureConfig.Collaborator.Maximum
		length1 := len(collaborators)
		length2 := len(invitations)
		if length1+length2 >= maximum {
			return ErrCollaboratorQuotaExceeded
		}
	}

	return nil
}

func (s *CollaboratorService) checkQuotaInAccept(ctx context.Context, appID string) error {
	appCtx, err := s.AppConfigs.ResolveContext(ctx, appID)
	if err != nil {
		return err
	}

	collaborators, err := s.ListCollaborators(ctx, appID)
	if err != nil {
		return err
	}

	if appCtx.Config.FeatureConfig.Collaborator.Maximum != nil {
		maximum := *appCtx.Config.FeatureConfig.Collaborator.Maximum
		length1 := len(collaborators)
		if length1 >= maximum {
			return ErrCollaboratorQuotaExceeded
		}
	}

	return nil
}

func scanCollaborator(scan db.Scanner) (*model.Collaborator, error) {
	c := &model.Collaborator{}

	err := scan.Scan(
		&c.ID,
		&c.AppID,
		&c.UserID,
		&c.CreatedAt,
		&c.Role,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCollaboratorNotFound
	} else if err != nil {
		return nil, err
	}

	return c, nil
}

func scanCollaboratorInvitation(scan db.Scanner) (*model.CollaboratorInvitation, error) {
	i := &model.CollaboratorInvitation{}

	err := scan.Scan(
		&i.ID,
		&i.AppID,
		&i.InvitedBy,
		&i.InviteeEmail,
		&i.Code,
		&i.CreatedAt,
		&i.ExpireAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCollaboratorInvitationNotFound
	} else if err != nil {
		return nil, err
	}

	return i, nil
}

func generateCollaboratorInvitationCode() string {
	code := rand.StringWithAlphabet(32, base32.Alphabet, rand.SecureRand)
	return code
}
