package application

import (
	"context"
	"errors"
	"slices"
	"testing"

	"PrService/src/internal/application/mocks"
	"PrService/src/internal/domain"

	"go.uber.org/mock/gomock"
)

func TestAssignReviewers_Table(t *testing.T) {
	author := domain.UserID("author")

	tests := []struct {
		name         string
		team         domain.Team
		authorID     domain.UserID
		wantExact    []domain.UserID
		wantLen      int
		maxReviewers int
	}{
		{
			name: "only author in team -> no reviewers",
			team: domain.Team{
				Name: "team",
				Members: []domain.TeamMember{
					{ID: author, Username: "author", IsActive: true},
				},
			},
			authorID:  author,
			wantExact: []domain.UserID{},
		},
		{
			name: "one active member besides author",
			team: domain.Team{
				Name: "team",
				Members: []domain.TeamMember{
					{ID: author, Username: "author", IsActive: true},
					{ID: "u2", Username: "u2", IsActive: true},
				},
			},
			authorID:  author,
			wantExact: []domain.UserID{"u2"},
		},
		{
			name: "inactive members are skipped",
			team: domain.Team{
				Name: "team",
				Members: []domain.TeamMember{
					{ID: author, Username: "author", IsActive: true},
					{ID: "active", Username: "active", IsActive: true},
					{ID: "inactive1", Username: "inactive1", IsActive: false},
					{ID: "inactive2", Username: "inactive2", IsActive: false},
				},
			},
			authorID:  author,
			wantExact: []domain.UserID{"active"},
		},
		{
			name: "more active members than PullRequestMaxReviewers",
			team: domain.Team{
				Name: "team",
				Members: []domain.TeamMember{
					{ID: author, Username: "author", IsActive: true},
					{ID: "u2", Username: "u2", IsActive: true},
					{ID: "u3", Username: "u3", IsActive: true},
					{ID: "u4", Username: "u4", IsActive: true},
				},
			},
			authorID:     author,
			wantLen:      domain.PullRequestMaxReviewers,
			maxReviewers: domain.PullRequestMaxReviewers,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := assignReviewers(tt.authorID, tt.team)

			if tt.wantExact != nil {
				if !slices.Equal(got, tt.wantExact) {
					t.Fatalf("expected %v, got %v", tt.wantExact, got)
				}
				return
			}

			if tt.wantLen != 0 && len(got) != tt.wantLen {
				t.Fatalf("expected len=%d, got %d", tt.wantLen, len(got))
			}

			if slices.Contains(got, tt.authorID) {
				t.Fatalf("reviewers must not contain author, got %v", got)
			}
			for _, id := range got {
				found := false
				for _, m := range tt.team.Members {
					if m.ID == id {
						found = true
						if !m.IsActive {
							t.Fatalf("inactive member %s selected as reviewer", id)
						}
						break
					}
				}
				if !found {
					t.Fatalf("reviewer %s is not a member of the team", id)
				}
			}
			if len(got) > domain.PullRequestMaxReviewers {
				t.Fatalf("len(got)=%d > PullRequestMaxReviewers", len(got))
			}
		})
	}
}

func TestReassignReviewers_Table(t *testing.T) {
	author := domain.UserID("author")
	oldRev := domain.UserID("rev-old")

	tests := []struct {
		name         string
		authorID     domain.UserID
		oldRevID     domain.UserID
		oldReviewers []domain.UserID
		team         domain.Team
		wantErr      error
		wantNew      []domain.UserID
		wantNewRevID domain.UserID
	}{
		{
			name:         "success: replace old reviewer with new candidate",
			authorID:     author,
			oldRevID:     oldRev,
			oldReviewers: []domain.UserID{oldRev},
			team: domain.Team{
				Name: "team",
				Members: []domain.TeamMember{
					{ID: author, IsActive: true},
					{ID: oldRev, IsActive: true},
					{ID: "rev-new", IsActive: true},
				},
			},
			wantErr:      nil,
			wantNew:      []domain.UserID{"rev-new"},
			wantNewRevID: "rev-new",
		},
		{
			name:         "no candidate in team -> ErrNoCandidate",
			authorID:     author,
			oldRevID:     oldRev,
			oldReviewers: []domain.UserID{oldRev},
			team: domain.Team{
				Name: "team",
				Members: []domain.TeamMember{
					{ID: author, IsActive: true},
					{ID: oldRev, IsActive: true},
				},
			},
			wantErr:      domain.ErrNoCandidate,
			wantNew:      []domain.UserID{},
			wantNewRevID: "",
		},
		{
			name:         "keep other reviewers and add new one",
			authorID:     author,
			oldRevID:     oldRev,
			oldReviewers: []domain.UserID{oldRev, "keep"},
			team: domain.Team{
				Name: "team",
				Members: []domain.TeamMember{
					{ID: author, IsActive: true},
					{ID: oldRev, IsActive: true},
					{ID: "keep", IsActive: true},
					{ID: "rev-new", IsActive: true},
				},
			},
			wantErr:      nil,
			wantNewRevID: "rev-new",
		},
		{
			name:         "no extra active candidate (only author, oldRev, and already assigned) -> ErrNoCandidate",
			authorID:     author,
			oldRevID:     oldRev,
			oldReviewers: []domain.UserID{oldRev, "keep"},
			team: domain.Team{
				Name: "team",
				Members: []domain.TeamMember{
					{ID: author, IsActive: true},
					{ID: oldRev, IsActive: true},
					{ID: "keep", IsActive: true},
				},
			},
			wantErr:      domain.ErrNoCandidate,
			wantNewRevID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReviewers, gotNewRev, err := reassignReviewers(tt.authorID, tt.oldRevID, tt.oldReviewers, tt.team)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantNew != nil {
				if !slices.Equal(gotReviewers, tt.wantNew) {
					t.Fatalf("expected reviewers %v, got %v", tt.wantNew, gotReviewers)
				}
			} else {
				if len(gotReviewers) != len(tt.oldReviewers) {
					t.Fatalf("len(gotReviewers)=%d, expected %d", len(gotReviewers), len(tt.oldReviewers))
				}
				if slices.Contains(gotReviewers, tt.oldRevID) {
					t.Fatalf("old reviewer %s must not be in new reviewers", tt.oldRevID)
				}
			}

			if gotNewRev != tt.wantNewRevID {
				t.Fatalf("expected new reviewer ID %s, got %s", tt.wantNewRevID, gotNewRev)
			}
		})
	}
}

func TestPullRequestService_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	teamRepo := mocks.NewMockTeamRepository(ctrl)
	txMgr := mocks.NewMockTxManager(ctrl)

	service := NewPullRequestService(prRepo, teamRepo, txMgr)

	ctx := context.Background()
	prID := domain.PullRequestID("pr-1")
	authorID := domain.UserID("author")

	team := &domain.Team{
		Name: "backend",
		Members: []domain.TeamMember{
			{ID: authorID, Username: "author", IsActive: true},
			{ID: "rev1", Username: "rev1", IsActive: true},
			{ID: "rev2", Username: "rev2", IsActive: true},
		},
	}

	txMgr.
		EXPECT().
		WithinTransaction(ctx, gomock.Any()).
		DoAndReturn(func(c context.Context, fn func(context.Context) error) error {
			return fn(c)
		})

	teamRepo.
		EXPECT().
		GetByUserID(gomock.Any(), authorID).
		Return(team, nil)

	prRepo.
		EXPECT().
		Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, pr *domain.PullRequest) error {
			if pr.ID != prID {
				t.Errorf("expected ID %s, got %s", prID, pr.ID)
			}
			if pr.AuthorID != authorID {
				t.Errorf("expected AuthorID %s, got %s", authorID, pr.AuthorID)
			}
			if pr.Name != "My PR" {
				t.Errorf("expected Name %s, got %s", "My PR", pr.Name)
			}
			if pr.Status != domain.PullRequestStatusOpen {
				t.Errorf("expected Status %s, got %s", domain.PullRequestStatusOpen, pr.Status)
			}
			if pr.CreatedAt == nil {
				t.Errorf("expected CreatedAt to be non-nil")
			}
			if pr.MergedAt != nil {
				t.Errorf("expected MergedAt to be nil")
			}
			if len(pr.AssignedReviewers) != domain.PullRequestMaxReviewers {
				t.Errorf("expected %d reviewers, got %d", domain.PullRequestMaxReviewers, len(pr.AssignedReviewers))
			}
			if slices.Contains(pr.AssignedReviewers, authorID) {
				t.Errorf("author must not be among reviewers")
			}
			return nil
		})

	pr, err := service.Create(ctx, prID, "My PR", authorID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pr == nil {
		t.Fatal("expected non-nil pull request")
	} else {
		if pr.ID != prID {
			t.Errorf("expected ID %s, got %s", prID, pr.ID)
		}
	}
}

func TestPullRequestService_Create_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	teamRepo := mocks.NewMockTeamRepository(ctrl)
	txMgr := mocks.NewMockTxManager(ctrl)

	service := NewPullRequestService(prRepo, teamRepo, txMgr)

	ctx := context.Background()
	prID := domain.PullRequestID("pr-1")
	authorID := domain.UserID("author")

	expectedErr := errors.New("tx error")

	txMgr.
		EXPECT().
		WithinTransaction(ctx, gomock.Any()).
		Return(expectedErr)

	pr, err := service.Create(ctx, prID, "My PR", authorID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
	if pr != nil {
		t.Fatalf("expected nil pull request, got %#v", pr)
	}
}

func TestPullRequestService_Create_GetTeamError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	teamRepo := mocks.NewMockTeamRepository(ctrl)
	txMgr := mocks.NewMockTxManager(ctrl)

	service := NewPullRequestService(prRepo, teamRepo, txMgr)

	ctx := context.Background()
	prID := domain.PullRequestID("pr-1")
	authorID := domain.UserID("author")

	expectedErr := errors.New("get team error")

	txMgr.
		EXPECT().
		WithinTransaction(ctx, gomock.Any()).
		DoAndReturn(func(c context.Context, fn func(context.Context) error) error {
			return fn(c)
		})

	teamRepo.
		EXPECT().
		GetByUserID(gomock.Any(), authorID).
		Return(nil, expectedErr)

	pr, err := service.Create(ctx, prID, "My PR", authorID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
	if pr != nil {
		t.Fatalf("expected nil pull request, got %#v", pr)
	}
}

func TestPullRequestService_Create_CreateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	teamRepo := mocks.NewMockTeamRepository(ctrl)
	txMgr := mocks.NewMockTxManager(ctrl)

	service := NewPullRequestService(prRepo, teamRepo, txMgr)

	ctx := context.Background()
	prID := domain.PullRequestID("pr-1")
	authorID := domain.UserID("author")

	team := &domain.Team{
		Name: "backend",
		Members: []domain.TeamMember{
			{ID: authorID, IsActive: true},
			{ID: "rev1", IsActive: true},
		},
	}

	expectedErr := errors.New("create pr error")

	txMgr.
		EXPECT().
		WithinTransaction(ctx, gomock.Any()).
		DoAndReturn(func(c context.Context, fn func(context.Context) error) error {
			return fn(c)
		})

	teamRepo.
		EXPECT().
		GetByUserID(gomock.Any(), authorID).
		Return(team, nil)

	prRepo.
		EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(expectedErr)

	pr, err := service.Create(ctx, prID, "My PR", authorID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
	if pr != nil {
		t.Fatalf("expected nil pull request, got %#v", pr)
	}
}

func TestPullRequestService_Merge_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	teamRepo := mocks.NewMockTeamRepository(ctrl)
	txMgr := mocks.NewMockTxManager(ctrl)

	service := NewPullRequestService(prRepo, teamRepo, txMgr)

	ctx := context.Background()
	prID := domain.PullRequestID("pr-1")

	pr := &domain.PullRequest{
		ID:     prID,
		Status: domain.PullRequestStatusOpen,
	}

	txMgr.
		EXPECT().
		WithinTransaction(ctx, gomock.Any()).
		DoAndReturn(func(c context.Context, fn func(context.Context) error) error {
			return fn(c)
		})

	prRepo.
		EXPECT().
		GetByID(gomock.Any(), prID).
		Return(pr, nil)

	prRepo.
		EXPECT().
		Update(gomock.Any(), pr).
		DoAndReturn(func(_ context.Context, updated *domain.PullRequest) error {
			if updated.Status != domain.PullRequestStatusMerged {
				t.Errorf("expected status MERGED, got %s", updated.Status)
			}
			if updated.MergedAt == nil {
				t.Errorf("expected MergedAt to be non-nil")
			}
			return nil
		})

	err := service.Merge(ctx, prID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPullRequestService_Merge_AlreadyMerged(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	teamRepo := mocks.NewMockTeamRepository(ctrl)
	txMgr := mocks.NewMockTxManager(ctrl)

	service := NewPullRequestService(prRepo, teamRepo, txMgr)

	ctx := context.Background()
	prID := domain.PullRequestID("pr-1")

	pr := &domain.PullRequest{
		ID:     prID,
		Status: domain.PullRequestStatusMerged,
	}

	txMgr.
		EXPECT().
		WithinTransaction(ctx, gomock.Any()).
		DoAndReturn(func(c context.Context, fn func(context.Context) error) error {
			return fn(c)
		})

	prRepo.
		EXPECT().
		GetByID(gomock.Any(), prID).
		Return(pr, nil)

	err := service.Merge(ctx, prID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPullRequestService_Merge_GetByIDError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	teamRepo := mocks.NewMockTeamRepository(ctrl)
	txMgr := mocks.NewMockTxManager(ctrl)

	service := NewPullRequestService(prRepo, teamRepo, txMgr)

	ctx := context.Background()
	prID := domain.PullRequestID("pr-1")

	expectedErr := errors.New("get pr error")

	txMgr.
		EXPECT().
		WithinTransaction(ctx, gomock.Any()).
		DoAndReturn(func(c context.Context, fn func(context.Context) error) error {
			return fn(c)
		})

	prRepo.
		EXPECT().
		GetByID(gomock.Any(), prID).
		Return(nil, expectedErr)

	err := service.Merge(ctx, prID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestPullRequestService_Merge_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	teamRepo := mocks.NewMockTeamRepository(ctrl)
	txMgr := mocks.NewMockTxManager(ctrl)

	service := NewPullRequestService(prRepo, teamRepo, txMgr)

	ctx := context.Background()
	prID := domain.PullRequestID("pr-1")

	expectedErr := errors.New("tx error")

	txMgr.
		EXPECT().
		WithinTransaction(ctx, gomock.Any()).
		Return(expectedErr)

	err := service.Merge(ctx, prID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestPullRequestService_Reassign_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	teamRepo := mocks.NewMockTeamRepository(ctrl)
	txMgr := mocks.NewMockTxManager(ctrl)

	service := NewPullRequestService(prRepo, teamRepo, txMgr)

	ctx := context.Background()
	prID := domain.PullRequestID("pr-1")
	authorID := domain.UserID("author")
	oldRevID := domain.UserID("rev-old")
	keepID := domain.UserID("rev-keep")
	newCandidateID := domain.UserID("rev-new")

	pr := &domain.PullRequest{
		ID:                prID,
		AuthorID:          authorID,
		Status:            domain.PullRequestStatusOpen,
		AssignedReviewers: []domain.UserID{oldRevID, keepID},
	}

	team := &domain.Team{
		Name: "backend",
		Members: []domain.TeamMember{
			{ID: authorID, IsActive: true},
			{ID: oldRevID, IsActive: true},
			{ID: keepID, IsActive: true},
			{ID: newCandidateID, IsActive: true},
		},
	}

	txMgr.
		EXPECT().
		WithinTransaction(ctx, gomock.Any()).
		DoAndReturn(func(c context.Context, fn func(context.Context) error) error {
			return fn(c)
		})

	prRepo.
		EXPECT().
		GetByID(gomock.Any(), prID).
		Return(pr, nil)

	teamRepo.
		EXPECT().
		GetByUserID(gomock.Any(), oldRevID).
		Return(team, nil)

	prRepo.
		EXPECT().
		Update(gomock.Any(), pr).
		DoAndReturn(func(_ context.Context, updated *domain.PullRequest) error {
			if updated.Status != domain.PullRequestStatusOpen {
				t.Errorf("expected status OPEN, got %s", updated.Status)
			}
			if len(updated.AssignedReviewers) != 2 {
				t.Errorf("expected 2 reviewers, got %d", len(updated.AssignedReviewers))
			}
			if slices.Contains(updated.AssignedReviewers, oldRevID) {
				t.Errorf("old reviewer must not be in AssignedReviewers")
			}
			if !slices.Contains(updated.AssignedReviewers, keepID) {
				t.Errorf("keep reviewer must remain in AssignedReviewers")
			}
			if !slices.Contains(updated.AssignedReviewers, newCandidateID) {
				t.Errorf("new candidate must be in AssignedReviewers")
			}
			return nil
		})

	updatedPR, newReviewer, err := service.Reassign(ctx, prID, oldRevID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updatedPR == nil {
		t.Fatal("expected non-nil pull request")
	}

	if newReviewer != newCandidateID {
		t.Fatalf("expected new reviewer %s, got %s", newCandidateID, newReviewer)
	}
}

func TestPullRequestService_Reassign_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	teamRepo := mocks.NewMockTeamRepository(ctrl)
	txMgr := mocks.NewMockTxManager(ctrl)

	service := NewPullRequestService(prRepo, teamRepo, txMgr)

	ctx := context.Background()
	prID := domain.PullRequestID("pr-1")
	oldRevID := domain.UserID("rev-old")

	expectedErr := errors.New("tx error")

	txMgr.
		EXPECT().
		WithinTransaction(ctx, gomock.Any()).
		Return(expectedErr)

	pr, newRev, err := service.Reassign(ctx, prID, oldRevID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
	if pr != nil {
		t.Fatalf("expected nil pull request, got %#v", pr)
	}
	if newRev != "" {
		t.Fatalf("expected empty new reviewer, got %s", newRev)
	}
}

func TestPullRequestService_Reassign_GetByIDError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	teamRepo := mocks.NewMockTeamRepository(ctrl)
	txMgr := mocks.NewMockTxManager(ctrl)

	service := NewPullRequestService(prRepo, teamRepo, txMgr)

	ctx := context.Background()
	prID := domain.PullRequestID("pr-1")
	oldRevID := domain.UserID("rev-old")

	expectedErr := errors.New("get pr error")

	txMgr.
		EXPECT().
		WithinTransaction(ctx, gomock.Any()).
		DoAndReturn(func(c context.Context, fn func(context.Context) error) error {
			return fn(c)
		})

	prRepo.
		EXPECT().
		GetByID(gomock.Any(), prID).
		Return(nil, expectedErr)

	pr, newRev, err := service.Reassign(ctx, prID, oldRevID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
	if pr != nil {
		t.Fatalf("expected nil pull request, got %#v", pr)
	}
	if newRev != "" {
		t.Fatalf("expected empty new reviewer, got %s", newRev)
	}
}

func TestPullRequestService_Reassign_MergedPR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	teamRepo := mocks.NewMockTeamRepository(ctrl)
	txMgr := mocks.NewMockTxManager(ctrl)

	service := NewPullRequestService(prRepo, teamRepo, txMgr)

	ctx := context.Background()
	prID := domain.PullRequestID("pr-1")
	oldRevID := domain.UserID("rev-old")

	pr := &domain.PullRequest{
		ID:                prID,
		Status:            domain.PullRequestStatusMerged,
		AssignedReviewers: []domain.UserID{oldRevID},
	}

	txMgr.
		EXPECT().
		WithinTransaction(ctx, gomock.Any()).
		DoAndReturn(func(c context.Context, fn func(context.Context) error) error {
			return fn(c)
		})

	prRepo.
		EXPECT().
		GetByID(gomock.Any(), prID).
		Return(pr, nil)

	prRes, newRev, err := service.Reassign(ctx, prID, oldRevID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrReassignMergedPullRequest) {
		t.Fatalf("expected ErrReassignMergedPullRequest, got %v", err)
	}
	if prRes != nil {
		t.Fatalf("expected nil pull request, got %#v", prRes)
	}
	if newRev != "" {
		t.Fatalf("expected empty new reviewer, got %s", newRev)
	}
}

func TestPullRequestService_Reassign_ReviewerNotAssigned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	teamRepo := mocks.NewMockTeamRepository(ctrl)
	txMgr := mocks.NewMockTxManager(ctrl)

	service := NewPullRequestService(prRepo, teamRepo, txMgr)

	ctx := context.Background()
	prID := domain.PullRequestID("pr-1")
	oldRevID := domain.UserID("rev-old")

	pr := &domain.PullRequest{
		ID:                prID,
		Status:            domain.PullRequestStatusOpen,
		AssignedReviewers: []domain.UserID{"other"},
	}

	txMgr.
		EXPECT().
		WithinTransaction(ctx, gomock.Any()).
		DoAndReturn(func(c context.Context, fn func(context.Context) error) error {
			return fn(c)
		})

	prRepo.
		EXPECT().
		GetByID(gomock.Any(), prID).
		Return(pr, nil)

	prRes, newRev, err := service.Reassign(ctx, prID, oldRevID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrReviewerIsNotAssigned) {
		t.Fatalf("expected ErrReviewerIsNotAssigned, got %v", err)
	}
	if prRes != nil {
		t.Fatalf("expected nil pull request, got %#v", prRes)
	}
	if newRev != "" {
		t.Fatalf("expected empty new reviewer, got %s", newRev)
	}
}

func TestPullRequestService_Reassign_GetTeamError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	teamRepo := mocks.NewMockTeamRepository(ctrl)
	txMgr := mocks.NewMockTxManager(ctrl)

	service := NewPullRequestService(prRepo, teamRepo, txMgr)

	ctx := context.Background()
	prID := domain.PullRequestID("pr-1")
	authorID := domain.UserID("author")
	oldRevID := domain.UserID("rev-old")

	pr := &domain.PullRequest{
		ID:                prID,
		AuthorID:          authorID,
		Status:            domain.PullRequestStatusOpen,
		AssignedReviewers: []domain.UserID{oldRevID},
	}

	expectedErr := errors.New("get team error")

	txMgr.
		EXPECT().
		WithinTransaction(ctx, gomock.Any()).
		DoAndReturn(func(c context.Context, fn func(context.Context) error) error {
			return fn(c)
		})

	prRepo.
		EXPECT().
		GetByID(gomock.Any(), prID).
		Return(pr, nil)

	teamRepo.
		EXPECT().
		GetByUserID(gomock.Any(), oldRevID).
		Return(nil, expectedErr)

	prRes, newRev, err := service.Reassign(ctx, prID, oldRevID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
	if prRes != nil {
		t.Fatalf("expected nil pull request, got %#v", prRes)
	}
	if newRev != "" {
		t.Fatalf("expected empty new reviewer, got %s", newRev)
	}
}

func TestPullRequestService_Reassign_NoCandidate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	teamRepo := mocks.NewMockTeamRepository(ctrl)
	txMgr := mocks.NewMockTxManager(ctrl)

	service := NewPullRequestService(prRepo, teamRepo, txMgr)

	ctx := context.Background()
	prID := domain.PullRequestID("pr-1")
	authorID := domain.UserID("author")
	oldRevID := domain.UserID("rev-old")

	pr := &domain.PullRequest{
		ID:                prID,
		AuthorID:          authorID,
		Status:            domain.PullRequestStatusOpen,
		AssignedReviewers: []domain.UserID{oldRevID},
	}

	team := &domain.Team{
		Name: "backend",
		Members: []domain.TeamMember{
			{ID: authorID, IsActive: true},
			{ID: oldRevID, IsActive: true},
		},
	}

	txMgr.
		EXPECT().
		WithinTransaction(ctx, gomock.Any()).
		DoAndReturn(func(c context.Context, fn func(context.Context) error) error {
			return fn(c)
		})

	prRepo.
		EXPECT().
		GetByID(gomock.Any(), prID).
		Return(pr, nil)

	teamRepo.
		EXPECT().
		GetByUserID(gomock.Any(), oldRevID).
		Return(team, nil)

	prRes, newRev, err := service.Reassign(ctx, prID, oldRevID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrNoCandidate) {
		t.Fatalf("expected ErrNoCandidate, got %v", err)
	}
	if prRes != nil {
		t.Fatalf("expected nil pull request, got %#v", prRes)
	}
	if newRev != "" {
		t.Fatalf("expected empty new reviewer, got %s", newRev)
	}
}

func TestPullRequestService_Reassign_UpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := mocks.NewMockPullRequestRepository(ctrl)
	teamRepo := mocks.NewMockTeamRepository(ctrl)
	txMgr := mocks.NewMockTxManager(ctrl)

	service := NewPullRequestService(prRepo, teamRepo, txMgr)

	ctx := context.Background()
	prID := domain.PullRequestID("pr-1")
	authorID := domain.UserID("author")
	oldRevID := domain.UserID("rev-old")

	pr := &domain.PullRequest{
		ID:                prID,
		AuthorID:          authorID,
		Status:            domain.PullRequestStatusOpen,
		AssignedReviewers: []domain.UserID{oldRevID},
	}

	team := &domain.Team{
		Name: "backend",
		Members: []domain.TeamMember{
			{ID: authorID, IsActive: true},
			{ID: oldRevID, IsActive: true},
			{ID: "rev-new", IsActive: true},
		},
	}

	expectedErr := errors.New("update error")

	txMgr.
		EXPECT().
		WithinTransaction(ctx, gomock.Any()).
		DoAndReturn(func(c context.Context, fn func(context.Context) error) error {
			return fn(c)
		})

	prRepo.
		EXPECT().
		GetByID(gomock.Any(), prID).
		Return(pr, nil)

	teamRepo.
		EXPECT().
		GetByUserID(gomock.Any(), oldRevID).
		Return(team, nil)

	prRepo.
		EXPECT().
		Update(gomock.Any(), pr).
		Return(expectedErr)

	prRes, newRev, err := service.Reassign(ctx, prID, oldRevID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
	if prRes != nil {
		t.Fatalf("expected nil pull request, got %#v", prRes)
	}
	if newRev != "" {
		t.Fatalf("expected empty new reviewer, got %s", newRev)
	}
}
