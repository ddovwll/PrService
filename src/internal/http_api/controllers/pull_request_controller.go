package controllers

import (
	"errors"
	"log/slog"
	"net/http"

	"PrService/src/internal/application/services"
	"PrService/src/internal/domain"
	"PrService/src/internal/http_api/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type PullRequestController struct {
	baseController
	pullRequestService *services.PullRequestService
}

func NewPullRequestController(
	pullRequestService *services.PullRequestService,
	validate *validator.Validate,
	logger *slog.Logger,
) *PullRequestController {
	return &PullRequestController{
		baseController:     newBaseController(validate, logger),
		pullRequestService: pullRequestService,
	}
}

func (c *PullRequestController) UseHandlers(r chi.Router) {
	r.Post("/pullRequest/create", c.create)
	r.Post("/pullRequest/merge", c.merge)
	r.Post("/pullRequest/reassign", c.reassign)
}

func (c *PullRequestController) create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.CreatePullRequestRequest
	if ok := c.decodeAndValidate(ctx, w, r, &req, "createPullRequestRequest"); !ok {
		return
	}

	pr, err := c.pullRequestService.Create(ctx,
		domain.PullRequestID(req.PullRequestID),
		req.PullRequestName,
		domain.UserID(req.AuthorID),
	)

	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.writeError(ctx, w, http.StatusNotFound,
				models.ErrorCodeNotFound,
				"resource not found",
				"user not found to create pull request",
				err,
				"pr_id", req.PullRequestID,
			)
			return
		}
		if errors.Is(err, domain.ErrTeamNotFound) {
			c.writeError(ctx, w, http.StatusNotFound,
				models.ErrorCodeNotFound,
				"resource not found",
				"team not found to create pull request",
				err,
				"pr_id", req.PullRequestID,
			)
			return
		}
		if errors.Is(err, domain.ErrPullRequestExists) {
			c.writeError(ctx, w, http.StatusConflict,
				models.ErrorCodePRExists,
				"PR id already exists",
				"PR id already exists",
				err,
				"pr_id", req.PullRequestID,
			)
			return
		}

		c.writeError(ctx, w, http.StatusInternalServerError,
			models.ErrorCodeInternalServer,
			"internal server error",
			"failed to create pull request",
			err,
			"pr_id", req.PullRequestID,
		)
		return
	}

	resp := models.MapToPullRequestEnvelopeResponse(*pr)
	c.writeJSON(ctx, w, http.StatusCreated, resp)
}

func (c *PullRequestController) merge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.MergePullRequestRequest
	if ok := c.decodeAndValidate(ctx, w, r, &req, "MergePullRequestRequest"); !ok {
		return
	}

	pr, err := c.pullRequestService.Merge(ctx, domain.PullRequestID(req.PullRequestID))

	if err != nil {
		if errors.Is(err, domain.ErrPullRequestNotFound) {
			c.writeError(ctx, w, http.StatusNotFound,
				models.ErrorCodeNotFound,
				"resource not found",
				"pr id not found",
				err,
				"pr_id", req.PullRequestID,
			)
			return
		}

		c.writeError(ctx, w, http.StatusInternalServerError,
			models.ErrorCodeInternalServer,
			"internal server error",
			"failed to merge pull request",
			err,
			"pr_id", req.PullRequestID,
		)
		return
	}

	resp := models.MapToPullRequestEnvelopeResponse(*pr)
	c.writeJSON(ctx, w, http.StatusOK, resp)
}

func (c *PullRequestController) reassign(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.ReassignPullRequestRequest
	if ok := c.decodeAndValidate(ctx, w, r, &req, "ReassignPullRequestRequest"); !ok {
		return
	}

	pr, replacedBy, err := c.pullRequestService.Reassign(ctx,
		domain.PullRequestID(req.PullRequestID),
		domain.UserID(req.OldUserID),
	)

	if err != nil {
		if errors.Is(err, domain.ErrPullRequestNotFound) {
			c.writeError(ctx, w, http.StatusNotFound,
				models.ErrorCodeNotFound,
				"resource not found",
				"pr not found",
				err,
				"pr_id", req.PullRequestID,
				"user_id", req.OldUserID,
			)
			return
		}
		if errors.Is(err, domain.ErrUserNotFound) {
			c.writeError(ctx, w, http.StatusNotFound,
				models.ErrorCodeNotFound,
				"resource not found",
				"user not found",
				err,
				"pr_id", req.PullRequestID,
				"user_id", req.OldUserID,
			)
			return
		}
		if errors.Is(err, domain.ErrReassignMergedPullRequest) {
			c.writeError(ctx, w, http.StatusConflict,
				models.ErrorCodePRMerged,
				"cannot reassign on merged PR",
				"pull request already merged",
				err,
				"pr_id", req.PullRequestID,
			)
			return
		}
		if errors.Is(err, domain.ErrReviewerIsNotAssigned) {
			c.writeError(ctx, w, http.StatusConflict,
				models.ErrorCodeNotAssigned,
				"user is not assigned on PR",
				"pull request reviewer is not assigned",
				err,
				"pr_id", req.PullRequestID,
				"user_id", req.OldUserID,
			)
			return
		}
		if errors.Is(err, domain.ErrNoCandidate) {
			c.writeError(ctx, w, http.StatusConflict,
				models.ErrorCodeNoCandidate,
				"no candidate to reassign review",
				"no candidate to reassign review",
				err,
				"pr_id", req.PullRequestID,
			)
			return
		}

		c.writeError(ctx, w, http.StatusInternalServerError,
			models.ErrorCodeInternalServer,
			"internal server error",
			"failed to reassign pull request",
			err,
			"pr_id", req.PullRequestID,
		)
		return
	}

	resp := models.MapToReassignPullRequestResponse(*pr, replacedBy)
	c.writeJSON(ctx, w, http.StatusOK, resp)
}
