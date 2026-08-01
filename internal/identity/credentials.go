package identity

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

type Credentials struct {
	InitialPassword string
	AgentAPIKey     string
	SessionID       string
}

func NewCredentials() (Credentials, error) {
	password, err := randomValue(24)
	if err != nil {
		return Credentials{}, err
	}
	apiKey, err := randomValue(32)
	if err != nil {
		return Credentials{}, err
	}
	sessionID, err := randomValue(32)
	if err != nil {
		return Credentials{}, err
	}
	return Credentials{
		InitialPassword: password,
		AgentAPIKey:     "thub_" + apiKey,
		SessionID:       sessionID,
	}, nil
}

func NewSessionID() (string, error) { return randomValue(32) }

func NewVerificationToken() (string, error) { return randomValue(32) }

func Digest(value string) []byte {
	sum := sha256.Sum256([]byte(value))
	return sum[:]
}

func randomValue(bytes int) (string, error) {
	raw := make([]byte, bytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
