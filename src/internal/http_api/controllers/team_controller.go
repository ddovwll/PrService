package controllers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"PrService/src/internal/application/services"
	"PrService/src/internal/domain"
	"PrService/src/internal/http_api/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type TeamController struct {
	baseController
	teamService *services.TeamService
}

func NewTeamController(
	teamService *services.TeamService,
	validate *validator.Validate,
	logger *slog.Logger,
) *TeamController {
	return &TeamController{
		baseController: newBaseController(validate, logger),
		teamService:    teamService,
	}
}

func (c *TeamController) UseHandlers(r chi.Router) {
	r.Post("/team/add", c.add)
	r.Get("/team/get", c.get)
}

func (c *TeamController) add(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.AddTeamRequest
	if ok := c.decodeAndValidate(ctx, w, r, &req, "addTeamRequest"); !ok {
		return
	}

	team := req.MapToDomain()
	createdTeam, err := c.teamService.Create(ctx, &team)
	if err != nil {
		if errors.Is(err, domain.ErrTeamAlreadyExists) {
			c.writeError(ctx, w, http.StatusBadRequest,
				models.ErrorCodeTeamExists,
				fmt.Sprintf("%s already exists", team.Name),
				"failed to write Create Team response",
				err,
				"team_name", req.TeamName,
			)
			return
		}

		c.writeError(ctx, w, http.StatusInternalServerError,
			models.ErrorCodeInternalServer,
			"internal server error",
			"failed to add team",
			err,
			"team_name", req.TeamName,
		)
		return
	}

	resp := models.MapToAddTeamResponse(*createdTeam)
	c.writeJSON(ctx, w, http.StatusCreated, resp)
}

func (c *TeamController) get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()
	teamName := q.Get("team_name")

	team, err := c.teamService.Get(ctx, domain.TeamName(teamName))
	if err != nil {
		if errors.Is(err, domain.ErrTeamNotFound) {
			c.writeError(ctx, w, http.StatusNotFound,
				models.ErrorCodeNotFound,
				fmt.Sprintf("%s already exists", team.Name),
				"failed to write Create Team response",
				err,
				"team_name", teamName,
			)
			return
		}

		c.writeError(ctx, w, http.StatusInternalServerError,
			models.ErrorCodeInternalServer,
			"internal server error",
			"failed to get team",
			err,
			"team_name", teamName,
		)
		return
	}

	resp := models.MapToTeamResponse(*team)
	c.writeJSON(ctx, w, http.StatusOK, resp)
}
