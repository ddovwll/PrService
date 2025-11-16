package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"PrService/src/internal/http_api/models"

	"github.com/go-playground/validator/v10"
)

func newHealthControllerForTest(t *testing.T) *HealthController {
	t.Helper()

	v := validator.New()
	logger := newTestLogger()

	return NewHealthController(v, logger)
}

func TestHealthController_Health(t *testing.T) {
	ctrl := newHealthControllerForTest(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	ctrl.health(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	//nolint:goconst // test file
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	var resp models.HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Status != "ok" {
		t.Fatalf("expected status \"ok\", got %q", resp.Status)
	}
}
