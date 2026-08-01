package identity

import "testing"

func TestNewCredentialsAreIndependentAndNonEmpty(t *testing.T) {
	first, err := NewCredentials()
	if err != nil {
		t.Fatalf("NewCredentials() error = %v", err)
	}
	second, err := NewCredentials()
	if err != nil {
		t.Fatalf("NewCredentials() error = %v", err)
	}
	if first.InitialPassword == "" || first.AgentAPIKey == "" || first.SessionID == "" {
		t.Fatal("NewCredentials() returned an empty credential")
	}
	if first.AgentAPIKey == first.InitialPassword || first.AgentAPIKey == first.SessionID {
		t.Fatal("NewCredentials() reused a credential across purposes")
	}
	if first.AgentAPIKey == second.AgentAPIKey {
		t.Fatal("NewCredentials() repeated API key")
	}
	if string(Digest(first.AgentAPIKey)) == first.AgentAPIKey {
		t.Fatal("Digest() returned plaintext")
	}
}
