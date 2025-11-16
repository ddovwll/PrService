package controllers

import (
	"log/slog"
	"net/http"

	"PrService/src/internal/http_api/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type HealthController struct {
	baseController
}

func NewHealthController(
	validate *validator.Validate,
	logger *slog.Logger,
) *HealthController {
	return &HealthController{
		baseController: newBaseController(validate, logger),
	}
}

func (c *HealthController) UseHandlers(r chi.Router) {
	r.Get("/health", c.health)
}

// health godoc
//
//	@Summary	Health check
//	@Tags		Health
//	@Produce	json
//	@Success	200	{object}	models.HealthResponse
//	@Router		/health [get]
func (c *HealthController) health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	resp := models.HealthResponse{
		Status: "ok",
	}
	c.writeJSON(ctx, w, http.StatusOK, resp)
}
