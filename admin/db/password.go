package db

import (
	"crypto/subtle"
	"fmt"
	"strings"

	"github.com/yincongcyincong/MuseBot/utils"
	"golang.org/x/crypto/bcrypt"
)

// DefaultAdminPassword is the well-known initial admin/admin password. The
// admin UI prompts for a password change after login, so the default is meant
// to be short-lived.
const DefaultAdminPassword = "admin"

// HasDefaultPassword reports whether the account still carries the default
// admin/admin password (bcrypt or legacy unsalted MD5 hash) so the UI can
// prompt for a change.
func HasDefaultPassword(u *User) bool {
	ok, _ := VerifyPassword(DefaultAdminPassword, u.Password)
	return ok
}

// HashPassword hashes a plaintext password with bcrypt.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password fail: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword reports whether the plaintext password matches the stored
// hash. Legacy unsalted MD5 hashes are still accepted so existing deployments
// keep working; isLegacy reports such a match so the caller can re-hash the
// password with bcrypt on the next successful login.
func VerifyPassword(password, hash string) (ok bool, isLegacy bool) {
	if isLegacyMD5Hash(hash) {
		legacy := utils.MD5(password)
		match := subtle.ConstantTimeCompare([]byte(legacy), []byte(hash)) == 1
		return match, match
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil, false
}

// isLegacyMD5Hash distinguishes old unsalted MD5 hashes (32 hex chars) from
// bcrypt hashes (prefixed with $2).
func isLegacyMD5Hash(hash string) bool {
	return len(hash) == 32 && !strings.HasPrefix(hash, "$2")
}
