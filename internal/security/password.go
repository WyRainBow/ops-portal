package security

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

// VerifyPassword supports:
// - bcrypt hashes (passlib default): $2a$... / $2b$...
// - passlib pbkdf2_sha256: $pbkdf2-sha256$29000$salt$hash
func VerifyPassword(plain string, hashed string) (bool, error) {
	if hashed == "" {
		return false, errors.New("empty hash")
	}
	if strings.HasPrefix(hashed, "$2a$") || strings.HasPrefix(hashed, "$2b$") || strings.HasPrefix(hashed, "$2y$") {
		err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
		if err == nil {
			return true, nil
		}
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}
	if strings.HasPrefix(hashed, "$pbkdf2-sha256$") {
		return verifyPasslibPBKDF2SHA256(plain, hashed)
	}
	// Unknown format: treat as mismatch.
	return false, nil
}

func verifyPasslibPBKDF2SHA256(plain, hashed string) (bool, error) {
	// Format: $pbkdf2-sha256$<rounds>$<salt>$<hash>
	parts := strings.Split(hashed, "$")
	if len(parts) < 5 {
		return false, fmt.Errorf("invalid pbkdf2-sha256 format")
	}
	roundsStr := parts[2]
	saltB64 := parts[3]
	hashB64 := parts[4]

	rounds, err := strconv.Atoi(roundsStr)
	if err != nil || rounds <= 0 {
		return false, fmt.Errorf("invalid pbkdf2 rounds")
	}

	salt, err := decodePasslibB64(saltB64)
	if err != nil {
		return false, err
	}
	want, err := decodePasslibB64(hashB64)
	if err != nil {
		return false, err
	}

	// passlib pbkdf2_sha256 uses dkLen=32.
	got := pbkdf2.Key([]byte(plain), salt, rounds, 32, sha256.New)
	if len(got) != len(want) {
		return false, nil
	}
	if subtle.ConstantTimeCompare(got, want) == 1 {
		return true, nil
	}
	return false, nil
}

func decodePasslibB64(s string) ([]byte, error) {
	// passlib uses URL-safe base64 without padding.
	s = strings.TrimSpace(s)
	if s == "" {
		return []byte{}, nil
	}
	// add padding to multiple of 4
	if m := len(s) % 4; m != 0 {
		s += strings.Repeat("=", 4-m)
	}
	// passlib uses ./ as possible? but for pbkdf2_sha256 it's typically urlsafe
	b, err := base64.URLEncoding.DecodeString(s)
	if err == nil {
		return b, nil
	}
	// fallback: standard
	b2, err2 := base64.StdEncoding.DecodeString(s)
	if err2 != nil {
		return nil, err
	}
	return b2, nil
}

