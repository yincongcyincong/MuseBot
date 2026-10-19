package db

import (
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/yincongcyincong/MuseBot/utils"
	"golang.org/x/crypto/bcrypt"
)

// DefaultAdminMD5Password is the unsalted MD5 hash of the well-known legacy
// password "admin". Accounts still holding this hash are reset on startup.
const DefaultAdminMD5Password = "21232f297a57a5a743894a0e4a801fc3"

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

// GenerateRandomPassword returns a random readable password printed once for
// the initial admin account.
func GenerateRandomPassword() (string, error) {
	const alphabet = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ23456789"
	const length = 12

	password := make([]byte, length)
	for i := range password {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", errors.New("generate random password fail: " + err.Error())
		}
		password[i] = alphabet[index.Int64()]
	}
	return string(password), nil
}
