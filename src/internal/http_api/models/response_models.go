package models

import (
	"PrService/src/internal/domain"
)

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

func CreateErrorResponse(code ErrorCode, mes string) ErrorResponse {
	return ErrorResponse{
		Error: ErrorBody{
			ErrorCode: code,
			Message:   mes,
		},
	}
}

type ErrorBody struct {
	ErrorCode ErrorCode `json:"code"`
	Message   string    `json:"message"`
}

type ErrorCode string

const (
	ErrorCodeTeamExists       ErrorCode = "TEAM_EXISTS"
	ErrorCodePRExists         ErrorCode = "PR_EXISTS"
	ErrorCodePRMerged         ErrorCode = "PR_MERGED"
	ErrorCodeNotAssigned      ErrorCode = "NOT_ASSIGNED"
	ErrorCodeNoCandidate      ErrorCode = "NO_CANDIDATE"
	ErrorCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrorCodeDecodeFailed     ErrorCode = "DECODE_FAILED"
	ErrorCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrorCodeInternalServer   ErrorCode = "INTERNAL_SERVER_ERROR"
)

type TeamMemberResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type TeamResponse struct {
	TeamName string               `json:"team_name"`
	Members  []TeamMemberResponse `json:"members"`
}

func MapToTeamResponse(team domain.Team) TeamResponse {
	members := make([]TeamMemberResponse, 0, len(team.Members))
	for _, member := range team.Members {
		members = append(members, TeamMemberResponse{
			UserID:   string(member.ID),
			Username: member.Username,
			IsActive: member.IsActive,
		})
	}

	return TeamResponse{
		TeamName: string(team.Name),
		Members:  members,
	}
}

type AddTeamResponse struct {
	Team TeamResponse `json:"team"`
}

func MapToAddTeamResponse(team domain.Team) AddTeamResponse {
	return AddTeamResponse{
		Team: MapToTeamResponse(team),
	}
}

type UserResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type SetUserIsActiveResponse struct {
	User UserResponse `json:"user"`
}

func MapToSetUserIsActiveResponse(user domain.User) SetUserIsActiveResponse {
	return SetUserIsActiveResponse{
		User: UserResponse{
			UserID:   string(user.ID),
			Username: user.Username,
			TeamName: string(user.TeamName),
			IsActive: user.IsActive,
		},
	}
}

type PullRequestShortResponse struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

type GetUserReviewsResponse struct {
	UserID       string                     `json:"user_id"`
	PullRequests []PullRequestShortResponse `json:"pull_requests"`
}

func MapToGetUserReviewsResponse(userID domain.UserID, prs []domain.PullRequest) GetUserReviewsResponse {
	prsResp := make([]PullRequestShortResponse, 0, len(prs))
	for _, pr := range prs {
		prsResp = append(prsResp, PullRequestShortResponse{
			PullRequestID:   string(pr.ID),
			PullRequestName: pr.Name,
			AuthorID:        string(pr.AuthorID),
			Status:          string(pr.Status),
		})
	}

	return GetUserReviewsResponse{
		UserID:       string(userID),
		PullRequests: prsResp,
	}
}

type PullRequestResponse struct {
	PullRequestID     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
	CreatedAt         *string  `json:"createdAt,omitempty"`
	MergedAt          *string  `json:"mergedAt,omitempty"`
}

func MapToPullRequestResponse(pr domain.PullRequest) PullRequestResponse {
	reviewers := make([]string, 0, len(pr.AssignedReviewers))
	for _, reviewer := range pr.AssignedReviewers {
		reviewers = append(reviewers, string(reviewer))
	}

	var createdAt *string
	if pr.CreatedAt != nil {
		createdAtFormatted := pr.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")
		createdAt = &createdAtFormatted
	}

	var mergedAt *string
	if pr.MergedAt != nil {
		mergedAtFormatted := pr.MergedAt.UTC().Format("2006-01-02T15:04:05Z")
		mergedAt = &mergedAtFormatted
	}

	return PullRequestResponse{
		PullRequestID:     string(pr.ID),
		PullRequestName:   pr.Name,
		AuthorID:          string(pr.AuthorID),
		Status:            string(pr.Status),
		AssignedReviewers: reviewers,
		CreatedAt:         createdAt,
		MergedAt:          mergedAt,
	}
}

type PullRequestEnvelopeResponse struct {
	PR PullRequestResponse `json:"pr"`
}

func MapToPullRequestEnvelopeResponse(pr domain.PullRequest) PullRequestEnvelopeResponse {
	return PullRequestEnvelopeResponse{
		PR: MapToPullRequestResponse(pr),
	}
}

type ReassignPullRequestResponse struct {
	PR         PullRequestResponse `json:"pr"`
	ReplacedBy string              `json:"replaced_by"`
}

func MapToReassignPullRequestResponse(pr domain.PullRequest, replacedBy domain.UserID) ReassignPullRequestResponse {
	return ReassignPullRequestResponse{
		PR:         MapToPullRequestResponse(pr),
		ReplacedBy: string(replacedBy),
	}
}

type HealthResponse struct {
	Status string `json:"status"`
}

type TeamStatsResponse struct {
	TeamName           string `json:"team_name"`
	MembersCount       int    `json:"members_count"`
	ActiveMembersCount int    `json:"active_members_count"`
	TotalPRs           int    `json:"total_prs"`
	OpenPRs            int    `json:"open_prs"`
	MergedPRs          int    `json:"merged_prs"`
	AvgTimeToMergeSec  int64  `json:"avg_time_to_merge_seconds"`
}

func MapToTeamStatsResponse(s domain.TeamStats) TeamStatsResponse {
	return TeamStatsResponse{
		TeamName:           string(s.TeamName),
		MembersCount:       s.MembersCount,
		ActiveMembersCount: s.ActiveMembersCount,
		TotalPRs:           s.TotalPRs,
		OpenPRs:            s.OpenPRs,
		MergedPRs:          s.MergedPRs,
		AvgTimeToMergeSec:  s.AvgTimeToMergeSec,
	}
}
