package application

import (
	"PrService/src/internal/domain"
	"context"
	"math/rand/v2"
	"slices"
	"time"
)

type PullRequestService struct {
	pullRequestRepository domain.PullRequestRepository
	teamRepository        domain.TeamRepository
	txManager             TxManager
}

func NewPullRequestService(
	pullRequestRepository domain.PullRequestRepository,
	teamRepository domain.TeamRepository,
	txManager TxManager,
) *PullRequestService {
	return &PullRequestService{
		pullRequestRepository: pullRequestRepository,
		teamRepository:        teamRepository,
		txManager:             txManager,
	}
}

func (s *PullRequestService) Create(
	ctx context.Context,
	id domain.PullRequestID,
	pullRequestName string,
	authorId domain.UserID,
) (*domain.PullRequest, error) {
	now := time.Now()
	// todo реализовать needMoreReviewers
	var pullRequest = &domain.PullRequest{
		ID:                id,
		Name:              pullRequestName,
		AuthorID:          authorId,
		Status:            domain.PullRequestStatusOpen,
		AssignedReviewers: []domain.UserID{},
		CreatedAt:         &now,
		MergedAt:          nil,
	}
	err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		team, err := s.teamRepository.GetByUserID(txCtx, authorId)
		if err != nil {
			return err
		}

		pullRequest.AssignedReviewers = assignReviewers(authorId, *team)

		if err := s.pullRequestRepository.Create(txCtx, pullRequest); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return pullRequest, nil
}

func assignReviewers(authorId domain.UserID, team domain.Team) []domain.UserID {
	members := append([]domain.TeamMember(nil), team.Members...)
	reviewers := make([]domain.UserID, 0)
	rand.Shuffle(len(members), func(i, j int) {
		members[i], members[j] = members[j], members[i]
	})

	for _, member := range members {
		if member.ID == authorId || !member.IsActive {
			continue
		}

		if len(reviewers) == domain.PullRequestMaxReviewers {
			break
		}

		reviewers = append(reviewers, member.ID)
	}

	return reviewers
}

func (s *PullRequestService) Merge(ctx context.Context, id domain.PullRequestID) error {
	err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		now := time.Now()

		pr, err := s.pullRequestRepository.GetByID(txCtx, id)
		if err != nil {
			return err
		}

		if pr.Status == domain.PullRequestStatusMerged {
			return nil
		}

		pr.Status = domain.PullRequestStatusMerged
		pr.MergedAt = &now

		return s.pullRequestRepository.Update(txCtx, pr)
	})

	return err
}

func (s *PullRequestService) Reassign(ctx context.Context, id domain.PullRequestID, oldRevId domain.UserID) (*domain.PullRequest, domain.UserID, error) {
	var pullRequest *domain.PullRequest
	var newReviewer domain.UserID
	err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		pr, err := s.pullRequestRepository.GetByID(txCtx, id)
		if err != nil {
			return err
		}

		if pr.Status == domain.PullRequestStatusMerged {
			return domain.ErrReassignMergedPullRequest
		}

		if !slices.Contains(pr.AssignedReviewers, oldRevId) {
			return domain.ErrReviewerIsNotAssigned
		}

		team, err := s.teamRepository.GetByUserID(txCtx, oldRevId)
		if err != nil {
			return err
		}

		newReviewers, newRevID, err := reassignReviewers(pr.AuthorID, oldRevId, pr.AssignedReviewers, *team)
		if err != nil {
			return err
		}
		newReviewer = newRevID

		pr.AssignedReviewers = newReviewers
		if err := s.pullRequestRepository.Update(txCtx, pr); err != nil {
			return err
		}

		pullRequest = pr

		return nil
	})

	if err != nil {
		return nil, "", err
	}

	return pullRequest, newReviewer, nil
}

func reassignReviewers(authorId, oldRevId domain.UserID, oldReviewers []domain.UserID, team domain.Team) ([]domain.UserID, domain.UserID, error) {
	members := append([]domain.TeamMember(nil), team.Members...)
	newReviewers := make([]domain.UserID, 0)
	rand.Shuffle(len(members), func(i, j int) {
		members[i], members[j] = members[j], members[i]
	})

	for _, reviewer := range oldReviewers {
		if reviewer == oldRevId {
			continue
		}

		newReviewers = append(newReviewers, reviewer)
	}

	var newReviewer domain.UserID
	for _, member := range members {
		if member.ID == authorId || !member.IsActive || member.ID == oldRevId {
			continue
		}

		if slices.Contains(newReviewers, member.ID) {
			continue
		}

		newReviewers = append(newReviewers, member.ID)
		newReviewer = member.ID
		break
	}

	if len(newReviewers) != len(oldReviewers) {
		return []domain.UserID{}, "", domain.ErrNoCandidate
	}

	return newReviewers, newReviewer, nil
}
