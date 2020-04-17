package forgotpassword

type Provider struct{}

// SendCode checks if loginID is an existing login ID.
// If not found, ErrLoginIDNotFound is returned.
// If the login ID is not of type email or phone, ErrUnsupportedLoginIDType is returned.
// Otherwise, a code is generated.
// The code expires after a specific time.
// The code becomes invalid if it is consumed.
// Finally the code is sent to the login ID asynchronously.
func (p *Provider) SendCode(loginID string) error {
	// TODO(forgotpassword)
	return nil
}

// ResetPassword consumes code and reset password to newPassword.
// If the code is invalid, ErrInvalidCode is returned.
// If the code is found but expired, ErrExpiredCode is returned.
// if the code is found but used, ErrUsedCode is returned.
// Otherwise, the password is reset to newPassword.
// newPassword is checked against the password policy so
// password policy error may also be returned.
func (p *Provider) ResetPassword(code string, newPassword string) error {
	// TODO(forgotpassword)
	return nil
}
