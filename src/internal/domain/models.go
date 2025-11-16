package domain

import "time"

type (
	UserID        string
	TeamName      string
	PullRequestID string
)

type PullRequestStatus string

const (
	PullRequestStatusOpen   PullRequestStatus = "OPEN"
	PullRequestStatusMerged PullRequestStatus = "MERGED"
	PullRequestMaxReviewers                   = 2
)

type TeamMember struct {
	ID       UserID
	Username string
	IsActive bool
}

func (m TeamMember) ToUser(teamName TeamName) User {
	return User{
		ID:       m.ID,
		Username: m.Username,
		TeamName: teamName,
		IsActive: m.IsActive,
	}
}

type Team struct {
	Name    TeamName
	Members []TeamMember
}

type TeamStats struct {
	TeamName           TeamName
	MembersCount       int
	ActiveMembersCount int
	TotalPRs           int
	OpenPRs            int
	MergedPRs          int
	AvgTimeToMergeSec  int64
}

type User struct {
	ID       UserID
	Username string
	TeamName TeamName
	IsActive bool
}

type PullRequest struct {
	ID                PullRequestID
	Name              string
	AuthorID          UserID
	Status            PullRequestStatus
	AssignedReviewers []UserID
	CreatedAt         *time.Time
	MergedAt          *time.Time
}
