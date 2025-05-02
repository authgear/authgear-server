package internal

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateMachineFingerprint generates a cryptographically secure random string
// to represent a machine.
// This approach works even in a containerized environment, unlike https://github.com/denisbrodbeck/machineid
func GenerateMachineFingerprint() string {
	// base64 encodes every 3 bytes into 4 bytes.
	// To avoid padding, the input length must be a multiple of 3.
	// 33 bytes = 264 bits > 256 bits
	randBytes := make([]byte, 33)
	_, err := rand.Read(randBytes)
	if err != nil {
		panic(fmt.Errorf("failed to generate random bytes: %w", err))
	}
	// Encode using base64url without padding
	return base64.RawURLEncoding.EncodeToString(randBytes)
}
