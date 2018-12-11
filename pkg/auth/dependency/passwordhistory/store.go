package passwordhistory

import (
	"time"
)

// Store encapsulates the interface of an Skygear Server connection to a container.
type Store interface {
	// CreatePasswordHistory create new password history.
	CreatePasswordHistory(userID string, hashedPassword []byte, loggedAt time.Time) error

	// GetPasswordHistory returns a slice of PasswordHistory of the given user
	//
	// If historySize is greater than 0, the returned slice contains history
	// of that size.
	// If historyDays is greater than 0, the returned slice contains history
	// up to now.
	//
	// If both historySize and historyDays are greater than 0, the returned slice
	// is the longer of the result.
	GetPasswordHistory(userID string, historySize, historyDays int) ([]PasswordHistory, error)

	// RemovePasswordHistory removes old password history.
	// It uses GetPasswordHistory to query active history and then purge old history.
	RemovePasswordHistory(userID string, historySize, historyDays int) error
}
