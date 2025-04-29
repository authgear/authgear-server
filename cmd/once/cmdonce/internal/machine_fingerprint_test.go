package internal

import (
	"encoding/base64"
	"regexp"
	"testing"
)

func FuzzGenerateMachineFingerprint(f *testing.F) {
	validChars := regexp.MustCompile(`^[A-Za-z0-9_-]*$`)

	f.Add(0)
	f.Fuzz(func(t *testing.T, _ int) {
		fingerprint := GenerateMachineFingerprint()

		// Test 1: Check if the fingerprint has the expected length
		// 33 bytes of random data encoded in base64url without padding should result in 44 characters
		expectedLength := 44
		if len(fingerprint) != expectedLength {
			t.Errorf("Expected fingerprint length to be %d, got %d", expectedLength, len(fingerprint))
		}

		// Test 2: Check if the fingerprint uses only valid base64url characters (A-Z, a-z, 0-9, -, _)
		if !validChars.MatchString(fingerprint) {
			t.Errorf("Fingerprint contains invalid characters: %s", fingerprint)
		}

		// Test 3: Check that the fingerprint does not end with padding characters (=)
		if fingerprint[len(fingerprint)-1] == '=' {
			t.Errorf("Fingerprint should not contain padding characters")
		}

		// Test 4: Verify that it can be decoded back to exactly 33 bytes
		decoded, err := base64.RawURLEncoding.DecodeString(fingerprint)
		if err != nil {
			t.Errorf("Failed to decode fingerprint: %v", err)
		}
		if len(decoded) != 33 {
			t.Errorf("Expected decoded length to be 33 bytes, got %d bytes", len(decoded))
		}
	})
}

func FuzzGenerateMachineFingerprintUniqueness(f *testing.F) {
	f.Add(0)
	f.Fuzz(func(t *testing.T, _ int) {
		// Generate two fingerprints and check they're different
		fp1 := GenerateMachineFingerprint()
		fp2 := GenerateMachineFingerprint()
		if fp1 == fp2 {
			t.Errorf("fingerprints should be unique: %v %v", fp1, fp2)
		}
	})
}
