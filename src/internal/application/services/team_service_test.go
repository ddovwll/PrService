package services

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"PrService/src/internal/application/mocks"
	"PrService/src/internal/domain"

	"go.uber.org/mock/gomock"
)

func TestTeamService_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	teamRepo := mocks.NewMockTeamRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	txManager := mocks.NewMockTxManager(ctrl)

	service := NewTeamService(teamRepo, userRepo, txManager)

	ctx := context.Background()
	teamName := domain.TeamName("backend")

	member1 := domain.TeamMember{
		ID:       domain.UserID("user-1"),
		Username: "alice",
		IsActive: true,
	}
	member2 := domain.TeamMember{
		ID:       domain.UserID("user-2"),
		Username: "bob",
		IsActive: false,
	}

	team := &domain.Team{
		Name:    teamName,
		Members: []domain.TeamMember{member1, member2},
	}

	expectedUsers := []domain.User{
		member1.ToUser(teamName),
		member2.ToUser(teamName),
	}

	txManager.
		EXPECT().
		WithinTransaction(ctx, gomock.Any()).
		DoAndReturn(func(c context.Context, fn func(context.Context) error) error {
			return fn(c)
		})

	teamRepo.
		EXPECT().
		Create(gomock.Any(), teamName).
		Return(nil)

	userRepo.
		EXPECT().
		UpsertBatch(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, users []domain.User) error {
			if !reflect.DeepEqual(users, expectedUsers) {
				t.Fatalf("unexpected users in UpsertBatch\nexpected: %#v\ngot:      %#v", expectedUsers, users)
			}
			return nil
		})

	result, err := service.Create(ctx, team)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil team")
	}

	if result != team {
		t.Errorf("expected service to return the same team pointer")
	}
}

func TestTeamService_Create_TeamCreateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	teamRepo := mocks.NewMockTeamRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	txManager := mocks.NewMockTxManager(ctrl)

	service := NewTeamService(teamRepo, userRepo, txManager)

	ctx := context.Background()

	team := &domain.Team{
		Name:    domain.TeamName("backend"),
		Members: []domain.TeamMember{},
	}

	expectedErr := errors.New("create team error")

	txManager.
		EXPECT().
		WithinTransaction(ctx, gomock.Any()).
		DoAndReturn(func(c context.Context, fn func(context.Context) error) error {
			return fn(c)
		})

	teamRepo.
		EXPECT().
		Create(gomock.Any(), team.Name).
		Return(expectedErr)

	result, err := service.Create(ctx, team)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	if result != nil {
		t.Fatalf("expected nil result on error, got %#v", result)
	}
}

func TestTeamService_Create_UpsertBatchError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	teamRepo := mocks.NewMockTeamRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	txManager := mocks.NewMockTxManager(ctrl)

	service := NewTeamService(teamRepo, userRepo, txManager)

	ctx := context.Background()
	teamName := domain.TeamName("backend")

	member := domain.TeamMember{
		ID:       domain.UserID("user-1"),
		Username: "alice",
		IsActive: true,
	}

	team := &domain.Team{
		Name:    teamName,
		Members: []domain.TeamMember{member},
	}

	expectedUsers := []domain.User{
		member.ToUser(teamName),
	}

	expectedErr := errors.New("upsert users error")

	txManager.
		EXPECT().
		WithinTransaction(ctx, gomock.Any()).
		DoAndReturn(func(c context.Context, fn func(context.Context) error) error {
			return fn(c)
		})

	teamRepo.
		EXPECT().
		Create(gomock.Any(), teamName).
		Return(nil)

	userRepo.
		EXPECT().
		UpsertBatch(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, users []domain.User) error {
			if !reflect.DeepEqual(users, expectedUsers) {
				t.Fatalf("unexpected users in UpsertBatch\nexpected: %#v\ngot:      %#v", expectedUsers, users)
			}
			return expectedErr
		})

	result, err := service.Create(ctx, team)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	if result != nil {
		t.Fatalf("expected nil result on error, got %#v", result)
	}
}

func TestTeamService_Create_TxManagerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	teamRepo := mocks.NewMockTeamRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	txManager := mocks.NewMockTxManager(ctrl)

	service := NewTeamService(teamRepo, userRepo, txManager)

	ctx := context.Background()
	team := &domain.Team{
		Name:    domain.TeamName("backend"),
		Members: []domain.TeamMember{},
	}

	expectedErr := errors.New("tx error")

	txManager.
		EXPECT().
		WithinTransaction(ctx, gomock.Any()).
		Return(expectedErr)

	result, err := service.Create(ctx, team)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	if result != nil {
		t.Fatalf("expected nil result on error, got %#v", result)
	}
}

func TestTeamService_Get_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	teamRepo := mocks.NewMockTeamRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	txManager := mocks.NewMockTxManager(ctrl)

	service := NewTeamService(teamRepo, userRepo, txManager)

	ctx := context.Background()
	name := domain.TeamName("backend")

	expectedTeam := &domain.Team{
		Name: name,
		Members: []domain.TeamMember{
			{
				ID:       domain.UserID("user-1"),
				Username: "alice",
				IsActive: true,
			},
		},
	}

	teamRepo.
		EXPECT().
		GetByName(ctx, name).
		Return(expectedTeam, nil)

	result, err := service.Get(ctx, name)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != expectedTeam {
		t.Fatalf("expected team %#v, got %#v", expectedTeam, result)
	}
}

func TestTeamService_Get_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	teamRepo := mocks.NewMockTeamRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	txManager := mocks.NewMockTxManager(ctrl)

	service := NewTeamService(teamRepo, userRepo, txManager)

	ctx := context.Background()
	name := domain.TeamName("backend")

	expectedErr := errors.New("get team error")

	teamRepo.
		EXPECT().
		GetByName(ctx, name).
		Return(nil, expectedErr)

	result, err := service.Get(ctx, name)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	if result != nil {
		t.Fatalf("expected nil result on error, got %#v", result)
	}
}

func TestTeamService_GetStats_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	teamRepo := mocks.NewMockTeamRepository(ctrl)

	svc := NewTeamService(teamRepo, nil, nil)

	ctx := context.Background()
	teamName := domain.TeamName("backend")

	expected := &domain.TeamStats{
		TeamName:           teamName,
		MembersCount:       3,
		ActiveMembersCount: 2,
		TotalPRs:           5,
		OpenPRs:            2,
		MergedPRs:          3,
		AvgTimeToMergeSec:  1234,
	}

	teamRepo.EXPECT().
		GetStats(ctx, teamName).
		Return(expected, nil)

	got, err := svc.GetStats(ctx, teamName)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != expected {
		t.Fatalf("expected %+v, got %+v", expected, got)
	}
}

func TestTeamService_GetStats_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	teamRepo := mocks.NewMockTeamRepository(ctrl)

	svc := NewTeamService(teamRepo, nil, nil)

	ctx := context.Background()
	teamName := domain.TeamName("unknown")

	teamRepo.EXPECT().
		GetStats(ctx, teamName).
		Return(nil, domain.ErrTeamNotFound)

	got, err := svc.GetStats(ctx, teamName)
	if got != nil {
		t.Fatalf("expected nil stats, got %+v", got)
	}
	if !errors.Is(err, domain.ErrTeamNotFound) {
		t.Fatalf("expected ErrTeamNotFound, got %v", err)
	}
}
