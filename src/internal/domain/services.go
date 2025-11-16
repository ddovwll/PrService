package domain

import "context"

type PullRequestService interface {
	Create(ctx context.Context, id PullRequestID, pullRequestName string, userID UserID) (*PullRequest, error)
	Merge(ctx context.Context, id PullRequestID) (*PullRequest, error)
	Reassign(ctx context.Context, id PullRequestID, oldRevID UserID) (*PullRequest, UserID, error)
}

type TeamService interface {
	Create(ctx context.Context, team *Team) (*Team, error)
	Get(ctx context.Context, name TeamName) (*Team, error)
	GetStats(ctx context.Context, name TeamName) (*TeamStats, error)
}

type UserService interface {
	SetIsActive(ctx context.Context, userID UserID, isActive bool) (*User, error)
	GetPrs(ctx context.Context, userID UserID) ([]PullRequest, error)
}
