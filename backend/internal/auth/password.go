package auth

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// MaxPasswordLen is bcrypt's effective input limit for GenerateFromPassword. bcrypt silently
// truncates inputs longer than this, so we reject them outright instead of authenticating a user
// against only the first 72 bytes of what they typed.
const MaxPasswordLen = 72

// ErrPasswordEmpty signals an empty / whitespace-only password.
var ErrPasswordEmpty = errors.New("password is empty")

// ErrPasswordTooLong signals the password exceeds bcrypt's 72-byte limit.
var ErrPasswordTooLong = errors.New("password too long")

// HashPassword returns a bcrypt hash. The only restrictions are non-empty and ≤ 72 bytes
// (bcrypt's hard limit). Length, case, symbols, and dictionary words are all up to the user.
func HashPassword(password string) (string, error) {
	raw := stringsTrimLen(password)
	if len(raw) == 0 {
		return "", ErrPasswordEmpty
	}
	if len(raw) > MaxPasswordLen {
		return "", ErrPasswordTooLong
	}
	b, err := bcrypt.GenerateFromPassword(raw, bcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CheckPassword compares plaintext with a bcrypt hash.
func CheckPassword(password, bcryptHash string) bool {
	raw := stringsTrimLen(password)
	if bcryptHash == "" || len(raw) == 0 {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(bcryptHash), raw) == nil
}

func stringsTrimLen(password string) []byte {
	s := strings.TrimSpace(password)
	return []byte(s)
}
