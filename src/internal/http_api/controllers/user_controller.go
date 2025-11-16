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

type UserController struct {
	baseController
	userService *services.UserService
}

func NewUserController(
	userService *services.UserService,
	validate *validator.Validate,
	logger *slog.Logger,
) *UserController {
	return &UserController{
		baseController: newBaseController(validate, logger),
		userService:    userService,
	}
}

func (c *UserController) UseHandlers(r chi.Router) {
	r.Post("/users/setIsActive", c.setIsActive)
	r.Get("/users/getReview", c.getReview)
}

// setIsActive godoc
// @Summary      Установить флаг активности пользователя
// @Description  Установить флаг активности пользователя по его идентификатору
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request  body      models.SetUserIsActiveRequest   true  "Set user is_active flag"
// @Success      200      {object}  models.SetUserIsActiveResponse
// @Failure      400      {object}  models.ErrorResponse  "invalid request body or validation failed"
// @Failure      404      {object}  models.ErrorResponse  "user not found"
// @Failure      500      {object}  models.ErrorResponse  "internal server error"
// @Router       /users/setIsActive [post]
func (c *UserController) setIsActive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.SetUserIsActiveRequest
	if ok := c.decodeAndValidate(ctx, w, r, &req, "setUserIsActiveRequest"); !ok {
		return
	}

	user, err := c.userService.SetIsActive(ctx, domain.UserID(req.UserID), req.IsActive)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.writeError(ctx, w, http.StatusNotFound,
				models.ErrorCodeNotFound,
				"resource not found",
				"user not found in SetIsActive",
				err,
				"user_id", req.UserID,
			)
			return
		}

		c.writeError(ctx, w, http.StatusInternalServerError,
			models.ErrorCodeInternalServer,
			"internal server error",
			"failed to set active status",
			err,
			"user_id", req.UserID,
		)
		return
	}

	resp := models.MapToSetUserIsActiveResponse(*user)
	c.writeJSON(ctx, w, http.StatusOK, resp)
}

// getReview godoc
// @Summary      Получить PR'ы, где пользователь назначен ревьювером
// @Description  Вернуть список PR'ов, в которых пользователь назначен ревьювером
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        user_id  query     string  true  "Идентификатор пользователя"
// @Success      200      {object}  models.GetUserReviewsResponse
// @Failure      400      {object}  models.ErrorResponse  "missing or invalid user_id"
// @Failure      500      {object}  models.ErrorResponse  "internal server error"
// @Router       /users/getReview [get]
func (c *UserController) getReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()
	userIDStr := q.Get("user_id")

	if userIDStr == "" {
		c.writeError(ctx, w, http.StatusBadRequest,
			models.ErrorCodeValidationFailed,
			"user_id is required",
			"missing user_id in query",
			nil,
		)
		return
	}

	prs, err := c.userService.GetPrs(ctx, domain.UserID(userIDStr))
	if err != nil {
		// если не нашлось — сейчас это ошибка сервиса; при необходимости можно
		// изменить реализацию сервиса, чтобы возвращать пустой список без ошибки.
		c.writeError(ctx, w, http.StatusInternalServerError,
			models.ErrorCodeInternalServer,
			"internal server error",
			"failed to get prs for user",
			err,
			"user_id", userIDStr,
		)
		return
	}

	resp := models.MapToGetUserReviewsResponse(domain.UserID(userIDStr), prs)
	c.writeJSON(ctx, w, http.StatusOK, resp)
}
