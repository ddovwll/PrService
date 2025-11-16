package controllers

import (
	"encoding/json"
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

func newPullRequestController(t *testing.T) (*PullRequestController, *mocks.MockPullRequestService) {
	t.Helper()

	ctrl := gomock.NewController(t)
	svc := mocks.NewMockPullRequestService(ctrl)

	validate := validator.New()
	logger := newTestLogger()

	c := NewPullRequestController(svc, validate, logger)

	return c, svc
}

func TestPullRequestController_Create_Success(t *testing.T) {
	c, svc := newPullRequestController(t)

	prID := domain.PullRequestID("pr-1")
	authorID := domain.UserID("u1")

	svc.
		EXPECT().
		Create(gomock.Any(), prID, "Test PR", authorID).
		Return(&domain.PullRequest{
			ID:                prID,
			Name:              "Test PR",
			AuthorID:          authorID,
			Status:            domain.PullRequestStatusOpen,
			AssignedReviewers: []domain.UserID{"u2", "u3"},
		}, nil)

	body := `{
		"pull_request_id": "pr-1",
		"pull_request_name": "Test PR",
		"author_id": "u1"
	}`

	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c.create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var resp models.PullRequestEnvelopeResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.PR.PullRequestID != string(prID) {
		t.Fatalf("expected pull_request_id=%s, got %s", prID, resp.PR.PullRequestID)
	}
	if resp.PR.AuthorID != string(authorID) {
		t.Fatalf("expected author_id=%s, got %s", authorID, resp.PR.AuthorID)
	}
	if resp.PR.Status != string(domain.PullRequestStatusOpen) {
		t.Fatalf("expected status=%s, got %s", domain.PullRequestStatusOpen, resp.PR.Status)
	}
}

func TestPullRequestController_Create_UserNotFound(t *testing.T) {
	c, svc := newPullRequestController(t)

	svc.
		EXPECT().
		Create(gomock.Any(), domain.PullRequestID("pr-2"), "Feature", domain.UserID("ghost")).
		Return(nil, domain.ErrUserNotFound)

	body := `{
		"pull_request_id": "pr-2",
		"pull_request_name": "Feature",
		"author_id": "ghost"
	}`

	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c.create(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusNotFound, rr.Code, rr.Body.String())
	}

	var errResp models.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}

	if errResp.Error.ErrorCode != models.ErrorCodeNotFound {
		t.Fatalf("expected error code %s, got %s", models.ErrorCodeNotFound, errResp.Error.ErrorCode)
	}
}

func TestPullRequestController_Create_TeamNotFound(t *testing.T) {
	c, svc := newPullRequestController(t)

	svc.
		EXPECT().
		Create(gomock.Any(), domain.PullRequestID("pr-3"), "Test", domain.UserID("u1")).
		Return(nil, domain.ErrTeamNotFound)

	body := `{
		"pull_request_id": "pr-3",
		"pull_request_name": "Test",
		"author_id": "u1"
	}`

	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c.create(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusNotFound, rr.Code, rr.Body.String())
	}
}

func TestPullRequestController_Create_AlreadyExists(t *testing.T) {
	c, svc := newPullRequestController(t)

	svc.
		EXPECT().
		Create(gomock.Any(), domain.PullRequestID("pr-4"), "Test", domain.UserID("u1")).
		Return(nil, domain.ErrPullRequestExists)

	body := `{
		"pull_request_id": "pr-4",
		"pull_request_name": "Test",
		"author_id": "u1"
	}`

	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c.create(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusConflict, rr.Code, rr.Body.String())
	}

	var errResp models.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}

	if errResp.Error.ErrorCode != models.ErrorCodePRExists {
		t.Fatalf("expected error code %s, got %s", models.ErrorCodePRExists, errResp.Error.ErrorCode)
	}
}

func TestPullRequestController_Merge_NotFound(t *testing.T) {
	c, svc := newPullRequestController(t)

	svc.
		EXPECT().
		Merge(gomock.Any(), domain.PullRequestID("unknown-pr")).
		Return(nil, domain.ErrPullRequestNotFound)

	body := `{"pull_request_id":"unknown-pr"}`

	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c.merge(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusNotFound, rr.Code, rr.Body.String())
	}

	var errResp models.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}

	if errResp.Error.ErrorCode != models.ErrorCodeNotFound {
		t.Fatalf("expected error code %s, got %s", models.ErrorCodeNotFound, errResp.Error.ErrorCode)
	}
}

func TestPullRequestController_Reassign_NoCandidate(t *testing.T) {
	c, svc := newPullRequestController(t)

	prID := domain.PullRequestID("pr-5")
	oldReviewer := domain.UserID("u2")

	svc.
		EXPECT().
		Reassign(gomock.Any(), prID, oldReviewer).
		Return(nil, domain.UserID(""), domain.ErrNoCandidate)

	body := `{
		"pull_request_id": "pr-5",
		"old_reviewer_id": "u2"
	}`

	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c.reassign(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusConflict, rr.Code, rr.Body.String())
	}

	var errResp models.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}

	if errResp.Error.ErrorCode != models.ErrorCodeNoCandidate {
		t.Fatalf("expected error code %s, got %s", models.ErrorCodeNoCandidate, errResp.Error.ErrorCode)
	}
}
