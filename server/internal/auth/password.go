package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Tunable parameters (good defaults for demos; you can increase memory/time for production)
const (
	argonTime    uint32 = 1          // iterations
	argonMemory  uint32 = 64 * 1024  // 64MB
	argonThreads uint8  = 4
	argonKeyLen  uint32 = 32
	saltLen             = 16
)

// HashPassword returns a versioned Argon2id hash string.
// Format: v=1$t=<time>$m=<memory>$p=<threads>$<base64url(salt)>$<base64url(hash)>
func HashPassword(plain string) (string, error) {
	if plain == "" {
		return "", errors.New("empty password")
	}

	// random salt
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("salt: %w", err)
	}

	sum := argon2.IDKey([]byte(plain), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	saltB64 := base64.RawURLEncoding.EncodeToString(salt)
	sumB64 := base64.RawURLEncoding.EncodeToString(sum)

	encoded := fmt.Sprintf("v=1$t=%d$m=%d$p=%d$%s$%s", argonTime, argonMemory, argonThreads, saltB64, sumB64)
	return encoded, nil
}

// VerifyPassword parses an encoded Argon2id hash and compares it to the provided password.
func VerifyPassword(plain, encoded string) bool {
	if plain == "" || encoded == "" {
		return false
	}

	parts := strings.Split(encoded, "$")
	// expect: v=1, t=.., m=.., p=.., salt, sum => 6 parts
	if len(parts) != 6 || !strings.HasPrefix(parts[0], "v=") {
		return false
	}
	// We ignore the parsed t/m/p in this demo and use the constants above.
	// If you want to support variable params, parse parts[1..3] and pass them into argon2.IDKey.

	salt, err := base64.RawURLEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	want, err := base64.RawURLEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	got := argon2.IDKey([]byte(plain), salt, argonTime, argonMemory, argonThreads, uint32(len(want)))

	// constant-time compare
	return subtle.ConstantTimeCompare(got, want) == 1
}
