package controllers

import (
	"encoding/json"
	"io"
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

func newTeamController(
	t *testing.T,
) (*TeamController, *mocks.MockTeamService) {
	t.Helper()

	ctrl := gomock.NewController(t)
	svc := mocks.NewMockTeamService(ctrl)

	validate := validator.New()
	logger := newTestLogger()

	c := NewTeamController(svc, validate, logger)

	return c, svc
}

func TestTeamController_Add_Success(t *testing.T) {
	c, svc := newTeamController(t)

	teamName := domain.TeamName("backend")

	svc.
		EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(&domain.Team{
			Name: teamName,
			Members: []domain.TeamMember{
				{ID: "u1", Username: "Alice", IsActive: true},
				{ID: "u2", Username: "Bob", IsActive: true},
			},
		}, nil)

	body := `{
		"team_name": "backend",
		"members": [
			{ "user_id": "u1", "username": "Alice", "is_active": true },
			{ "user_id": "u2", "username": "Bob",   "is_active": true }
		]
	}`

	req := httptest.NewRequest(http.MethodPost, "/team/add", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c.add(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var resp models.AddTeamResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal AddTeamResponse: %v", err)
	}

	if resp.Team.TeamName != string(teamName) {
		t.Fatalf("expected team_name=%s, got %s", teamName, resp.Team.TeamName)
	}

	if len(resp.Team.Members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(resp.Team.Members))
	}
}

func TestTeamController_Add_TeamAlreadyExists(t *testing.T) {
	c, svc := newTeamController(t)

	svc.
		EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(nil, domain.ErrTeamAlreadyExists)

	body := `{
		"team_name": "backend",
		"members": []
	}`

	req := httptest.NewRequest(http.MethodPost, "/team/add", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c.add(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}

	var errResp models.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal ErrorResponse: %v", err)
	}

	if errResp.Error.ErrorCode != models.ErrorCodeTeamExists {
		t.Fatalf("expected error code %s, got %s", models.ErrorCodeTeamExists, errResp.Error.ErrorCode)
	}
}

func TestTeamController_Get_Success(t *testing.T) {
	c, svc := newTeamController(t)

	teamName := domain.TeamName("backend")

	svc.
		EXPECT().
		Get(gomock.Any(), teamName).
		Return(&domain.Team{
			Name: teamName,
			Members: []domain.TeamMember{
				{ID: "u1", Username: "Alice", IsActive: true},
				{ID: "u2", Username: "Bob", IsActive: false},
			},
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=backend", nil)
	rr := httptest.NewRecorder()

	c.get(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp models.TeamResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal TeamResponse: %v", err)
	}

	if resp.TeamName != string(teamName) {
		t.Fatalf("expected team_name=%s, got %s", teamName, resp.TeamName)
	}
	if len(resp.Members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(resp.Members))
	}
}

func TestTeamController_Get_NotFound(t *testing.T) {
	c, svc := newTeamController(t)

	svc.
		EXPECT().
		Get(gomock.Any(), domain.TeamName("unknown")).
		Return(nil, domain.ErrTeamNotFound)

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=unknown", nil)
	rr := httptest.NewRecorder()

	c.get(rr, req)

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

func TestTeamController_Stats_Success(t *testing.T) {
	c, svc := newTeamController(t)

	teamName := "backend"
	domainTeamName := domain.TeamName(teamName)

	expectedStats := &domain.TeamStats{
		TeamName:           domainTeamName,
		MembersCount:       3,
		ActiveMembersCount: 2,
		TotalPRs:           5,
		OpenPRs:            2,
		MergedPRs:          3,
		AvgTimeToMergeSec:  1234,
	}

	svc.
		EXPECT().
		GetStats(gomock.Any(), domainTeamName).
		Return(expectedStats, nil)

	req := httptest.NewRequest(http.MethodGet, "/team/stats?team_name="+teamName, nil)
	rr := httptest.NewRecorder()

	c.stats(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	bodyBytes, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	var resp models.TeamStatsResponse
	if err := json.Unmarshal(bodyBytes, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.TeamName != teamName {
		t.Fatalf("expected team_name=%q, got %q", teamName, resp.TeamName)
	}
	if resp.MembersCount != expectedStats.MembersCount ||
		resp.ActiveMembersCount != expectedStats.ActiveMembersCount ||
		resp.TotalPRs != expectedStats.TotalPRs ||
		resp.OpenPRs != expectedStats.OpenPRs ||
		resp.MergedPRs != expectedStats.MergedPRs ||
		resp.AvgTimeToMergeSec != expectedStats.AvgTimeToMergeSec {
		t.Fatalf("unexpected stats response: %+v", resp)
	}
}

func TestTeamController_Stats_MissingTeamName(t *testing.T) {
	c, svc := newTeamController(t)

	svc.
		EXPECT().
		GetStats(gomock.Any(), gomock.Any()).
		Times(0)

	req := httptest.NewRequest(http.MethodGet, "/team/stats", nil)
	rr := httptest.NewRecorder()

	c.stats(rr, req)

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
}

func TestTeamController_Stats_TeamNotFound(t *testing.T) {
	c, svc := newTeamController(t)

	teamName := "ghost"
	domainTeamName := domain.TeamName(teamName)

	svc.
		EXPECT().
		GetStats(gomock.Any(), domainTeamName).
		Return(nil, domain.ErrTeamNotFound)

	req := httptest.NewRequest(http.MethodGet, "/team/stats?team_name="+teamName, nil)
	rr := httptest.NewRecorder()

	c.stats(rr, req)

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
