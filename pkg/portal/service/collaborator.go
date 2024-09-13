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
	SelfDirector(actorUserID string, usage Usage) (func(*http.Request), error)
}

type CollaboratorAppConfigService interface {
	ResolveContext(appID string) (*config.AppContext, error)
}

type CollaboratorService struct {
	Context     context.Context
	Clock       clock.Clock
	SQLBuilder  *globaldb.SQLBuilder
	SQLExecutor *globaldb.SQLExecutor

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

func (s *CollaboratorService) ListCollaborators(appID string) ([]*model.Collaborator, error) {
	q := s.selectCollaborator().Where("app_id = ?", appID)
	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cs []*model.Collaborator
	for rows.Next() {
		c, err := scanCollaborator(rows)
		if err != nil {
			return nil, err
		}
		cs = append(cs, c)
	}

	return cs, nil
}

func (s *CollaboratorService) ListCollaboratorsByUser(userID string) ([]*model.Collaborator, error) {
	q := s.selectCollaborator().Where("user_id = ?", userID)
	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cs []*model.Collaborator
	for rows.Next() {
		c, err := scanCollaborator(rows)
		if err != nil {
			return nil, err
		}
		cs = append(cs, c)
	}

	return cs, nil
}

func (s *CollaboratorService) GetProjectOwnerCount(userID string) (int, error) {
	q := s.SQLBuilder.Select("count(1)").
		From(s.SQLBuilder.TableName("_portal_app_collaborator")).
		Where("user_id = ?", userID).
		Where("role = ?", string(model.CollaboratorRoleOwner))

	row, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return 0, err
	}

	var count int
	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *CollaboratorService) GetManyProjectOwnerCount(userIDs []string) ([]int, error) {
	q := s.SQLBuilder.Select(
		"user_id",
		"count(1)",
	).
		From(s.SQLBuilder.TableName("_portal_app_collaborator")).
		Where("user_id = ANY(?)", pq.Array(userIDs)).
		GroupBy("user_id", "role").
		Having("role = ?", string(model.CollaboratorRoleOwner))

	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[string]int)
	for rows.Next() {
		var userID string
		var count int
		err = rows.Scan(&userID, &count)
		if err != nil {
			return nil, err
		}
		m[userID] = count
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

func (s *CollaboratorService) CreateCollaborator(c *model.Collaborator) error {
	err := s.deleteExpiredInvitations()
	if err != nil {
		return err
	}

	_, err = s.SQLExecutor.ExecWith(s.SQLBuilder.
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

func (s *CollaboratorService) GetCollaborator(id string) (*model.Collaborator, error) {
	q := s.selectCollaborator().Where("id = ?", id)
	row, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}
	return scanCollaborator(row)
}

func (s *CollaboratorService) GetManyCollaborators(ids []string) ([]*model.Collaborator, error) {
	q := s.selectCollaborator().Where("id = ANY (?)", pq.Array(ids))
	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cs []*model.Collaborator
	for rows.Next() {
		c, err := scanCollaborator(rows)
		if err != nil {
			return nil, err
		}
		cs = append(cs, c)
	}

	return cs, nil
}

func (s *CollaboratorService) GetCollaboratorByAppAndUser(appID string, userID string) (*model.Collaborator, error) {
	q := s.selectCollaborator().Where("app_id = ? AND user_id = ?", appID, userID)
	row, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}
	return scanCollaborator(row)
}

func (s *CollaboratorService) DeleteCollaborator(c *model.Collaborator) error {
	sessionInfo := session.GetValidSessionInfo(s.Context)
	if c.UserID == sessionInfo.UserID {
		return ErrCollaboratorSelfDeletion
	}

	err := s.deleteExpiredInvitations()
	if err != nil {
		return err
	}

	_, err = s.SQLExecutor.ExecWith(s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_portal_app_collaborator")).
		Where("id = ?", c.ID),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *CollaboratorService) GetManyInvitations(ids []string) ([]*model.CollaboratorInvitation, error) {
	q := s.selectCollaboratorInvitation().Where("id = ANY (?)", pq.Array(ids))
	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var is []*model.CollaboratorInvitation
	for rows.Next() {
		i, err := scanCollaboratorInvitation(rows)
		if err != nil {
			return nil, err
		}
		is = append(is, i)
	}

	return is, nil
}

func (s *CollaboratorService) ListInvitations(appID string) ([]*model.CollaboratorInvitation, error) {
	now := s.Clock.NowUTC()
	q := s.selectCollaboratorInvitation().Where("app_id = ? AND expire_at > ?", appID, now)
	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var is []*model.CollaboratorInvitation
	for rows.Next() {
		i, err := scanCollaboratorInvitation(rows)
		if err != nil {
			return nil, err
		}
		is = append(is, i)
	}

	return is, nil
}

func (s *CollaboratorService) SendInvitation(
	appID string,
	inviteeEmail string,
) (*model.CollaboratorInvitation, error) {
	sessionInfo := session.GetValidSessionInfo(s.Context)
	invitedBy := sessionInfo.UserID

	err := s.checkQuotaInSend(appID)
	if err != nil {
		return nil, err
	}

	// TODO(collaborator): Ideally we should prevent sending invitation to existing collaborator.
	// However, this is not harmful to not have it.
	// The collaborator will receive the invitation and they cannot accept it because
	// we have database constraint to enforce this invariant.
	// If Admin API have getUserByClaim, then we can detect this condition here.

	// Check if the invitee has a pending invitation already.
	invitations, err := s.ListInvitations(appID)
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

	err = s.deleteExpiredInvitations()
	if err != nil {
		return nil, err
	}

	err = s.createCollaboratorInvitation(i)
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

func (s *CollaboratorService) GetInvitation(id string) (*model.CollaboratorInvitation, error) {
	q := s.selectCollaboratorInvitation().Where("id = ?", id)
	row, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}
	return scanCollaboratorInvitation(row)
}

func (s *CollaboratorService) GetInvitationWithCode(code string) (*model.CollaboratorInvitation, error) {
	now := s.Clock.NowUTC()
	q := s.selectCollaboratorInvitation().Where("code = ? AND expire_at > ?", code, now)
	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var is []*model.CollaboratorInvitation
	for rows.Next() {
		i, err := scanCollaboratorInvitation(rows)
		if err != nil {
			return nil, err
		}
		is = append(is, i)
	}

	if len(is) <= 0 {
		return nil, ErrCollaboratorInvitationInvalidCode
	}

	return is[0], nil
}

func (s *CollaboratorService) DeleteInvitation(i *model.CollaboratorInvitation) error {
	err := s.deleteExpiredInvitations()
	if err != nil {
		return err
	}

	_, err = s.SQLExecutor.ExecWith(s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_portal_app_collaborator_invitation")).
		Where("id = ?", i.ID),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *CollaboratorService) AcceptInvitation(code string) (*model.Collaborator, error) {
	actorID := session.GetValidSessionInfo(s.Context).UserID

	invitation, err := s.GetInvitationWithCode(code)
	if err != nil {
		return nil, err
	}

	err = s.checkQuotaInAccept(invitation.AppID)
	if err != nil {
		return nil, err
	}

	err = s.CheckInviteeEmail(invitation, actorID)
	if err != nil {
		return nil, err
	}

	err = s.DeleteInvitation(invitation)
	if err != nil {
		return nil, err
	}

	_, err = s.GetCollaboratorByAppAndUser(invitation.AppID, actorID)
	if err == nil {
		return nil, ErrCollaboratorDuplicate
	}

	if !errors.Is(err, ErrCollaboratorNotFound) {
		return nil, err
	}

	collaborator := s.NewCollaborator(invitation.AppID, actorID, model.CollaboratorRoleEditor)
	err = s.CreateCollaborator(collaborator)
	if err != nil {
		return nil, err
	}

	return collaborator, nil
}

func (s *CollaboratorService) deleteExpiredInvitations() error {
	now := s.Clock.NowUTC()
	_, err := s.SQLExecutor.ExecWith(s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_portal_app_collaborator_invitation")).
		Where("expire_at <= ?", now),
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *CollaboratorService) createCollaboratorInvitation(i *model.CollaboratorInvitation) error {
	_, err := s.SQLExecutor.ExecWith(s.SQLBuilder.
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

func (s *CollaboratorService) CheckInviteeEmail(i *model.CollaboratorInvitation, actorID string) error {
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

	director, err := s.AdminAPI.SelfDirector(actorID, UsageInternal)
	if err != nil {
		return err
	}

	director(r)

	result, err := graphqlutil.HTTPDo(r, params)
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

func (s *CollaboratorService) checkQuotaInSend(appID string) error {
	appCtx, err := s.AppConfigs.ResolveContext(appID)
	if err != nil {
		return err
	}

	collaborators, err := s.ListCollaborators(appID)
	if err != nil {
		return err
	}

	invitations, err := s.ListInvitations(appID)
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

func (s *CollaboratorService) checkQuotaInAccept(appID string) error {
	appCtx, err := s.AppConfigs.ResolveContext(appID)
	if err != nil {
		return err
	}

	collaborators, err := s.ListCollaborators(appID)
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
