package domain

import "context"

type TeamRepository interface {
	Create(ctx context.Context, name TeamName) error
	GetByName(ctx context.Context, name TeamName) (*Team, error)
	GetByUserID(ctx context.Context, userID UserID) (*Team, error)
}
type UserRepository interface {
	UpsertBatch(ctx context.Context, users []User) error
	GetByID(ctx context.Context, id UserID) (*User, error)
	Update(ctx context.Context, user *User) error
}

type PullRequestRepository interface {
	Create(ctx context.Context, pr *PullRequest) error
	GetByID(ctx context.Context, id PullRequestID) (*PullRequest, error)
	ListByReviewer(ctx context.Context, reviewerID UserID) ([]PullRequest, error)
	Update(ctx context.Context, pr *PullRequest) error
}
