package authflowv2

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsV2HTML = template.RegisterHTML(
	"web/authflowv2/settings.html",
	handlerwebapp.SettingsComponents...,
)

func ConfigureAuthflowV2SettingsRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET", "POST").
		WithPathPattern(SettingsV2RouteSettings)
}

type SettingsAccountDeletionViewModel struct {
	AccountDeletionAllowed bool
}

// SettingsLoginIDRowsViewModel decides which login ID rows are links. A row is
// not a link when the page it leads to has nothing to do.
type SettingsLoginIDRowsViewModel struct {
	// LoginIDViewActionable is keyed by identity ID.
	LoginIDViewActionable map[string]bool
	// LoginIDKeyHasActions is keyed by login ID key.
	LoginIDKeyHasActions map[string]bool
}

type AuthflowV2SettingsHandler struct {
	Database                 *appdb.Handle
	ControllerFactory        handlerwebapp.ControllerFactory
	BaseViewModel            *viewmodels.BaseViewModeler
	AuthenticationViewModel  *viewmodels.AuthenticationViewModeler
	SettingsViewModel        *viewmodels.SettingsViewModeler
	SettingsProfileViewModel *viewmodels.SettingsProfileViewModeler
	Identities               SettingsIdentityService
	Renderer                 handlerwebapp.Renderer
	AccountDeletion          *config.AccountDeletionConfig
	LoginIDConfig            *config.LoginIDConfig
	VerificationConfig       *config.VerificationConfig
	Verification             SettingsVerificationService
}

func (h *AuthflowV2SettingsHandler) GetData(ctx context.Context, r *http.Request, rw http.ResponseWriter) (map[string]any, error) {
	userID := session.GetUserID(ctx)

	data := map[string]any{}

	// BaseViewModel
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	// SettingsViewModel
	viewModelPtr, err := h.SettingsViewModel.ViewModel(ctx, *userID)
	if err != nil {
		return nil, err
	}
	viewmodels.Embed(data, *viewModelPtr)

	// SettingsProfileViewModel
	profileViewModelPtr, err := h.SettingsProfileViewModel.ViewModel(ctx, *userID)
	if err != nil {
		return nil, err
	}
	viewmodels.Embed(data, *profileViewModelPtr)

	// Identity - Part 1
	candidates, err := h.Identities.ListCandidates(ctx, *userID)
	if err != nil {
		return nil, err
	}
	authenticationViewModel := h.AuthenticationViewModel.NewWithCandidates(candidates, r.Form)
	viewmodels.Embed(data, authenticationViewModel)

	allIdentities, err := h.Identities.ListByUser(ctx, *userID)
	if err != nil {
		return nil, err
	}
	viewActionable, err := LoginIDViewActions(ctx, allIdentities, h.Verification, h.LoginIDConfig, h.VerificationConfig)
	if err != nil {
		return nil, err
	}
	identityCountByKey := map[string]int{}
	keyHasActions := map[string]bool{}
	for _, info := range allIdentities {
		if info.Type != model.IdentityTypeLoginID {
			continue
		}
		key := info.LoginID.LoginIDKey
		identityCountByKey[key]++
		if viewActionable[info.ID] {
			keyHasActions[key] = true
		}
	}
	for _, c := range candidates {
		key, _ := c[identity.CandidateKeyLoginIDKey].(string)
		if key == "" {
			continue
		}
		createDisabled, _ := c[identity.CandidateKeyCreateDisabled].(bool)
		isUsername := c[identity.CandidateKeyLoginIDType] == string(model.LoginIDKeyTypeUsername)
		// A username key holds one identity only.
		if !createDisabled && !(isUsername && identityCountByKey[key] > 0) {
			keyHasActions[key] = true
		}
	}
	viewmodels.Embed(data, SettingsLoginIDRowsViewModel{
		LoginIDViewActionable: viewActionable,
		LoginIDKeyHasActions:  keyHasActions,
	})

	// Account Deletion
	accountDeletionViewModel := SettingsAccountDeletionViewModel{
		AccountDeletionAllowed: h.AccountDeletion.ScheduledByEndUserEnabled,
	}
	viewmodels.Embed(data, accountDeletionViewModel)

	return data, nil
}

func (h *AuthflowV2SettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx(r.Context())

	ctrl.Get(func(ctx context.Context) error {
		var data map[string]any
		err := h.Database.WithTx(ctx, func(ctx context.Context) error {
			data, err = h.GetData(ctx, r, w)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsV2HTML, data)
		return nil
	})
}
