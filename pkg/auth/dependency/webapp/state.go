package webapp

import (
	"net/url"

	corerand "github.com/skygeario/skygear-server/pkg/core/rand"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

var (
	stateIDAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	stateIDLength   = 32
)

// State management
// The webapp adopts the Post/Redirect/Get pattern.
//
// In this pattern, we cannot persist state directly by rendering
// hidden form fields in the response of POST request.
//
// Here we use a simple approach to work around this limitation.
//
// In the first POST request of a flow, a state object is created.
// The x_sid query parameter in the URL identities a state object.
//
// This approach does not use cookie at all.
type State struct {
	// ID is a cryptographically random string.
	ID string `json:"id"`
	// Form is encoded url.Values. It stores hidden form fields.
	Form string `json:"form"`
	// Error is either reset to nil or set to non-nil in every POST request.
	Error *skyerr.APIError `json:"error"`
	// AnonymousUserID is the ID of anonymous user during promotion flow.
	AnonymousUserID string `json:"anonymous_user_id,omitempty"`
}

func NewState() *State {
	return &State{
		ID: corerand.StringWithAlphabet(stateIDLength, stateIDAlphabet, corerand.SecureRand),
	}
}

func (s *State) SetForm(form url.Values) {
	s.Form = form.Encode()
}

func (s *State) SetError(err error) {
	s.Error = skyerr.AsAPIError(err)
}

// Restore merge s.Form into form.
// In case of conflict, form wins.
// This allows state update.
func (s *State) Restore(form url.Values) (err error) {
	thisForm, err := url.ParseQuery(s.Form)
	if err != nil {
		return
	}
	for name := range thisForm {
		_, ok := form[name]
		if ok {
			// Do not overwrite form
			continue
		}
		form.Set(name, thisForm.Get(name))
	}
	return
}
