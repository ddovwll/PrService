package application

import (
	"context"

	"PrService/src/internal/domain"
)

type UserService struct {
	userRepository        domain.UserRepository
	pullRequestRepository domain.PullRequestRepository
}

func NewUserService(
	userRepository domain.UserRepository,
	pullRequestRepository domain.PullRequestRepository,
) *UserService {
	return &UserService{
		userRepository:        userRepository,
		pullRequestRepository: pullRequestRepository,
	}
}

func (s *UserService) SetIsActive(ctx context.Context, userID domain.UserID, isActive bool) (*domain.User, error) {
	user, err := s.userRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.IsActive = isActive
	err = s.userRepository.Update(ctx, user)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetPrs(ctx context.Context, userID domain.UserID) ([]domain.PullRequest, error) {
	prs, err := s.pullRequestRepository.ListByReviewer(ctx, userID)
	if err != nil {
		return nil, err
	}

	return prs, nil
}
