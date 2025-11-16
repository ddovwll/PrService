package controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"PrService/src/internal/http_api/models"

	"github.com/go-playground/validator/v10"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func newTestBaseController(t *testing.T) *baseController {
	t.Helper()

	v := validator.New()
	logger := newTestLogger()

	bc := newBaseController(v, logger)
	return &bc
}

func TestBaseController_WriteJSON_Success(t *testing.T) {
	bc := newTestBaseController(t)

	rr := httptest.NewRecorder()
	ctx := context.Background()

	payload := map[string]string{"key": "value"}

	bc.writeJSON(ctx, rr, http.StatusCreated, payload)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	var got map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}

	if got["key"] != "value" {
		t.Fatalf("expected key=value, got %v", got)
	}
}

func TestBaseController_WriteJSON_NilPayload(t *testing.T) {
	bc := newTestBaseController(t)

	rr := httptest.NewRecorder()
	ctx := context.Background()

	bc.writeJSON(ctx, rr, http.StatusNoContent, nil)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}

	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	if rr.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", rr.Body.String())
	}
}

func TestBaseController_WriteError_4xx(t *testing.T) {
	bc := newTestBaseController(t)

	rr := httptest.NewRecorder()
	ctx := context.Background()

	bc.writeError(ctx, rr,
		http.StatusNotFound,
		models.ErrorCodeNotFound,
		"resource not found",
		"some log message",
		nil,
		"field", "value",
	)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}

	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	var errResp models.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}

	if errResp.Error.ErrorCode != models.ErrorCodeNotFound {
		t.Fatalf("expected error code %s, got %s", models.ErrorCodeNotFound, errResp.Error.ErrorCode)
	}
	if errResp.Error.Message != "resource not found" {
		t.Fatalf("expected message %q, got %q", "resource not found", errResp.Error.Message)
	}
}

func TestBaseController_WriteError_5xx(t *testing.T) {
	bc := newTestBaseController(t)

	rr := httptest.NewRecorder()
	ctx := context.Background()

	bc.writeError(ctx, rr,
		http.StatusInternalServerError,
		models.ErrorCodeInternalServer,
		"internal server error",
		"some internal error",
		nil,
	)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	var errResp models.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}

	if errResp.Error.ErrorCode != models.ErrorCodeInternalServer {
		t.Fatalf("expected error code %s, got %s", models.ErrorCodeInternalServer, errResp.Error.ErrorCode)
	}
}

func TestBaseController_DecodeAndValidate_Success(t *testing.T) {
	bc := newTestBaseController(t)

	type testRequest struct {
		Name string `json:"name" validate:"required"`
	}

	body := `{"name":"alice"}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	ctx := context.Background()

	var dst testRequest

	ok := bc.decodeAndValidate(ctx, rr, req, &dst, "testRequest")
	if !ok {
		t.Fatalf("expected decodeAndValidate to return true")
	}

	if dst.Name != "alice" {
		t.Fatalf("expected Name=alice, got %q", dst.Name)
	}

	if rr.Body.Len() != 0 {
		t.Fatalf("expected empty body for successful decodeAndValidate, got %q", rr.Body.String())
	}
}

func TestBaseController_DecodeAndValidate_DecodeError(t *testing.T) {
	bc := newTestBaseController(t)

	type testRequest struct {
		Name string `json:"name" validate:"required"`
	}

	body := `{"name":`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	ctx := context.Background()

	var dst testRequest

	ok := bc.decodeAndValidate(ctx, rr, req, &dst, "testRequest")
	if ok {
		t.Fatalf("expected decodeAndValidate to return false on invalid JSON")
	}

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}

	var errResp models.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}

	if errResp.Error.ErrorCode != models.ErrorCodeDecodeFailed {
		t.Fatalf("expected error code %s, got %s", models.ErrorCodeDecodeFailed, errResp.Error.ErrorCode)
	}
	if errResp.Error.Message != "invalid request body" {
		t.Fatalf("expected message %q, got %q", "invalid request body", errResp.Error.Message)
	}
}

func TestBaseController_DecodeAndValidate_ValidationError(t *testing.T) {
	bc := newTestBaseController(t)

	type testRequest struct {
		Name string `json:"name" validate:"required"`
	}

	body := `{}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	ctx := context.Background()

	var dst testRequest

	ok := bc.decodeAndValidate(ctx, rr, req, &dst, "testRequest")
	if ok {
		t.Fatalf("expected decodeAndValidate to return false on validation error")
	}

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}

	var errResp models.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}

	if errResp.Error.ErrorCode != models.ErrorCodeValidationFailed {
		t.Fatalf("expected error code %s, got %s", models.ErrorCodeValidationFailed, errResp.Error.ErrorCode)
	}
	if errResp.Error.Message != "validation failed" {
		t.Fatalf("expected message %q, got %q", "validation failed", errResp.Error.Message)
	}
}
