package domain

import "errors"

var (
	ErrTeamNotFound              = errors.New("team not found")
	ErrTeamAlreadyExists         = errors.New("team already exists")
	ErrUserNotFound              = errors.New("user not found")
	ErrPullRequestNotFound       = errors.New("pull request not found")
	ErrPullRequestExists         = errors.New("pull request already exists")
	ErrReassignMergedPullRequest = errors.New("cannot reassign on merged PR")
	ErrNoCandidate               = errors.New("cannot find candidate")
	ErrReviewerIsNotAssigned     = errors.New("reviewer is not assigned")
)
