package models

import "PrService/src/internal/domain"

type AddTeamRequest struct {
	TeamName string              `json:"team_name" validate:"required"`
	Members  []TeamMemberRequest `json:"members" validate:"required,dive"`
}

type TeamMemberRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	Username string `json:"username" validate:"required"`
	IsActive bool   `json:"is_active"`
}

func (team AddTeamRequest) MapToDomain() domain.Team {
	members := make([]domain.TeamMember, 0, len(team.Members))
	for _, member := range team.Members {
		members = append(members, domain.TeamMember{
			ID:       domain.UserID(member.UserID),
			Username: member.Username,
			IsActive: member.IsActive,
		})
	}

	return domain.Team{
		Name:    domain.TeamName(team.TeamName),
		Members: members,
	}
}

type GetTeamRequest struct {
	TeamName string `validate:"required"`
}

type SetUserIsActiveRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	IsActive bool   `json:"is_active"`
}

type CreatePullRequestRequest struct {
	PullRequestID   string `json:"pull_request_id" validate:"required"`
	PullRequestName string `json:"pull_request_name" validate:"required"`
	AuthorID        string `json:"author_id" validate:"required"`
}

type MergePullRequestRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required"`
}

type ReassignPullRequestRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required"`
	OldUserID     string `json:"old_reviewer_id" validate:"required"`
}
