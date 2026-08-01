package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsUnsafeOriginRequiresTheConfiguredOriginForStateChanges(t *testing.T) {
	allowed := "https://travelinghub.example"
	request := httptest.NewRequest(http.MethodPost, "/v1/web/change-password", nil)
	if !isUnsafeOrigin(request, allowed) {
		t.Fatal("state-changing request without Origin was accepted")
	}

	request.Header.Set("Origin", allowed)
	if isUnsafeOrigin(request, allowed) {
		t.Fatal("state-changing request from configured Origin was rejected")
	}

	request.Header.Set("Origin", "https://attacker.example")
	if !isUnsafeOrigin(request, allowed) {
		t.Fatal("state-changing cross-origin request was accepted")
	}

	readRequest := httptest.NewRequest(http.MethodGet, "/v1/game", nil)
	if isUnsafeOrigin(readRequest, allowed) {
		t.Fatal("safe read request was rejected")
	}
}
