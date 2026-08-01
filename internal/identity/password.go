package identity

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    uint32 = 1
	argonMemory  uint32 = 64 * 1024
	argonThreads uint8  = 4
	argonKeyLen  uint32 = 32
)

func HashPassword(password string) (string, error) {
	if len(password) < 12 {
		return "", fmt.Errorf("password must be at least 12 characters")
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("read password salt: %w", err)
	}
	digest := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf(
		"argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(digest),
	), nil
}

func VerifyPassword(encoded, password string) bool {
	memory, timeCost, threads, salt, expected, ok := parseHash(encoded)
	if !ok {
		return false
	}
	actual := argon2.IDKey([]byte(password), salt, timeCost, memory, threads, uint32(len(expected)))
	return subtle.ConstantTimeCompare(actual, expected) == 1
}

func parseHash(encoded string) (uint32, uint32, uint8, []byte, []byte, bool) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 5 || parts[0] != "argon2id" || parts[1] != "v=19" {
		return 0, 0, 0, nil, nil, false
	}
	values := map[string]string{}
	for _, part := range strings.Split(parts[2], ",") {
		pair := strings.SplitN(part, "=", 2)
		if len(pair) != 2 {
			return 0, 0, 0, nil, nil, false
		}
		values[pair[0]] = pair[1]
	}
	memory64, err := strconv.ParseUint(values["m"], 10, 32)
	if err != nil || memory64 == 0 {
		return 0, 0, 0, nil, nil, false
	}
	time64, err := strconv.ParseUint(values["t"], 10, 32)
	if err != nil || time64 == 0 {
		return 0, 0, 0, nil, nil, false
	}
	threads64, err := strconv.ParseUint(values["p"], 10, 8)
	if err != nil || threads64 == 0 {
		return 0, 0, 0, nil, nil, false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil || len(salt) < 16 {
		return 0, 0, 0, nil, nil, false
	}
	digest, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(digest) < 16 {
		return 0, 0, 0, nil, nil, false
	}
	return uint32(memory64), uint32(time64), uint8(threads64), salt, digest, true
}
