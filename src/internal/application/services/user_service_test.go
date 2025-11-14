package services

import (
	"context"
	"errors"
	"testing"

	"PrService/src/internal/application/mocks"
	"PrService/src/internal/domain"

	"go.uber.org/mock/gomock"
)

func TestUserService_SetIsActive_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	prRepo := mocks.NewMockPullRequestRepository(ctrl)

	service := NewUserService(userRepo, prRepo)

	ctx := context.Background()
	var userID domain.UserID

	user := &domain.User{IsActive: false}

	userRepo.
		EXPECT().
		GetByID(ctx, userID).
		Return(user, nil)

	userRepo.
		EXPECT().
		Update(ctx, user).
		DoAndReturn(func(_ context.Context, u *domain.User) error {
			if !u.IsActive {
				t.Errorf("expected user.IsActive to be true before Update call")
			}
			return nil
		})

	updatedUser, err := service.SetIsActive(ctx, userID, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updatedUser == nil {
		t.Fatal("expected non-nil user")
	} else {
		if !updatedUser.IsActive {
			t.Errorf("expected IsActive == true, got %v", updatedUser.IsActive)
		}
	}

	if updatedUser != user {
		t.Errorf("expected to return same user pointer from repository")
	}
}

func TestUserService_SetIsActive_GetByIDError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	prRepo := mocks.NewMockPullRequestRepository(ctrl)

	service := NewUserService(userRepo, prRepo)

	ctx := context.Background()
	var userID domain.UserID

	expectedErr := errors.New("get user error")

	userRepo.
		EXPECT().
		GetByID(ctx, userID).
		Return(nil, expectedErr)

	user, err := service.SetIsActive(ctx, userID, true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	if user != nil {
		t.Fatalf("expected nil user on error, got %#v", user)
	}
}

func TestUserService_SetIsActive_UpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	prRepo := mocks.NewMockPullRequestRepository(ctrl)

	service := NewUserService(userRepo, prRepo)

	ctx := context.Background()
	var userID domain.UserID

	user := &domain.User{IsActive: false}
	expectedErr := errors.New("update user error")

	userRepo.
		EXPECT().
		GetByID(ctx, userID).
		Return(user, nil)

	userRepo.
		EXPECT().
		Update(ctx, user).
		Return(expectedErr)

	updatedUser, err := service.SetIsActive(ctx, userID, true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	if updatedUser != nil {
		t.Fatalf("expected nil user on error, got %#v", updatedUser)
	}
}

func TestUserService_GetPrs_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	prRepo := mocks.NewMockPullRequestRepository(ctrl)

	service := NewUserService(userRepo, prRepo)

	ctx := context.Background()
	var userID domain.UserID

	prsFromRepo := []domain.PullRequest{
		{},
		{},
	}

	prRepo.
		EXPECT().
		ListByReviewer(ctx, userID).
		Return(prsFromRepo, nil)

	prs, err := service.GetPrs(ctx, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if prs == nil {
		t.Fatal("expected non-nil slice")
	}

	if len(prs) != len(prsFromRepo) {
		t.Fatalf("expected %d PRs, got %d", len(prsFromRepo), len(prs))
	}
}

func TestUserService_GetPrs_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	prRepo := mocks.NewMockPullRequestRepository(ctrl)

	service := NewUserService(userRepo, prRepo)

	ctx := context.Background()
	var userID domain.UserID

	expectedErr := errors.New("list prs error")

	prRepo.
		EXPECT().
		ListByReviewer(ctx, userID).
		Return(nil, expectedErr)

	prs, err := service.GetPrs(ctx, userID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	if prs != nil {
		t.Fatalf("expected nil slice on error, got %#v", prs)
	}
}
