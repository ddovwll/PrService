package controllers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"PrService/src/internal/domain"
	"PrService/src/internal/http_api/mocks"
	"PrService/src/internal/http_api/models"

	"github.com/go-playground/validator/v10"
	"go.uber.org/mock/gomock"
)

func newUserController(
	t *testing.T,
) (*UserController, *mocks.MockUserService) {
	t.Helper()

	ctrl := gomock.NewController(t)
	svc := mocks.NewMockUserService(ctrl)

	validate := validator.New()
	logger := slog.New(slog.DiscardHandler)

	c := NewUserController(svc, validate, logger)

	return c, svc
}

func TestUserController_SetIsActive_Success(t *testing.T) {
	c, svc := newUserController(t)

	userID := domain.UserID("u1")

	svc.
		EXPECT().
		SetIsActive(gomock.Any(), userID, false).
		Return(&domain.User{
			ID:       userID,
			Username: "Alice",
			TeamName: "backend",
			IsActive: false,
		}, nil)

	body := `{"user_id":"u1","is_active":false}`

	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c.setIsActive(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp models.SetUserIsActiveResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal SetUserIsActiveResponse: %v", err)
	}

	if resp.User.UserID != string(userID) {
		t.Fatalf("expected user_id=%s, got %s", userID, resp.User.UserID)
	}
	if resp.User.IsActive {
		t.Fatalf("expected is_active=false in response, got true")
	}
}

func TestUserController_SetIsActive_UserNotFound(t *testing.T) {
	c, svc := newUserController(t)

	svc.
		EXPECT().
		SetIsActive(gomock.Any(), domain.UserID("ghost"), true).
		Return(nil, domain.ErrUserNotFound)

	body := `{"user_id":"ghost","is_active":true}`

	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c.setIsActive(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusNotFound, rr.Code, rr.Body.String())
	}

	var errResp models.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal ErrorResponse: %v", err)
	}

	if errResp.Error.ErrorCode != models.ErrorCodeNotFound {
		t.Fatalf("expected error code %s, got %s", models.ErrorCodeNotFound, errResp.Error.ErrorCode)
	}
}

func TestUserController_GetReview_MissingUserID(t *testing.T) {
	c, _ := newUserController(t)

	req := httptest.NewRequest(http.MethodGet, "/users/getReview", nil)
	rr := httptest.NewRecorder()

	c.getReview(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}

	var errResp models.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal ErrorResponse: %v", err)
	}

	if errResp.Error.ErrorCode != models.ErrorCodeValidationFailed {
		t.Fatalf("expected error code %s, got %s", models.ErrorCodeValidationFailed, errResp.Error.ErrorCode)
	}
}

func TestUserController_GetReview_Success(t *testing.T) {
	c, svc := newUserController(t)

	userID := domain.UserID("u2")

	svc.
		EXPECT().
		GetPrs(gomock.Any(), userID).
		Return([]domain.PullRequest{
			{
				ID:       "pr-1",
				Name:     "Add search",
				AuthorID: "u1",
				Status:   domain.PullRequestStatusOpen,
			},
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u2", nil)
	rr := httptest.NewRecorder()

	c.getReview(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp models.GetUserReviewsResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal GetUserReviewsResponse: %v", err)
	}

	if resp.UserID != string(userID) {
		t.Fatalf("expected user_id=%s, got %s", userID, resp.UserID)
	}
	if len(resp.PullRequests) != 1 {
		t.Fatalf("expected 1 pull request, got %d", len(resp.PullRequests))
	}
	if resp.PullRequests[0].PullRequestID != "pr-1" {
		t.Fatalf("expected pull_request_id=pr-1, got %s", resp.PullRequests[0].PullRequestID)
	}
}

func TestUserController_GetReview_ServiceError(t *testing.T) {
	c, svc := newUserController(t)

	userID := domain.UserID("u2")

	svc.
		EXPECT().
		GetPrs(gomock.Any(), userID).
		Return(nil, errors.New("service error"))

	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u2", nil)
	rr := httptest.NewRecorder()

	c.getReview(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusInternalServerError, rr.Code, rr.Body.String())
	}
}
