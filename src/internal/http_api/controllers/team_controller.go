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

// add godoc
// @Summary      Создать команду с участниками
// @Description  Создать команду с участниками (создаёт/обновляет пользователей)
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Param        request  body      models.AddTeamRequest       true  "Team with members"
// @Success      201      {object}  models.AddTeamResponse
// @Failure      400      {object}  models.ErrorResponse  "team already exists or invalid request"
// @Failure      500      {object}  models.ErrorResponse  "internal server error"
// @Router       /team/add [post]
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
				fmt.Sprintf("%s already exists", req.TeamName),
				"team already exists",
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

// get godoc
// @Summary      Получить команду с участниками
// @Description  Получить команду и список её участников
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Param        team_name  query     string  true  "Уникальное имя команды"
// @Success      200        {object}  models.TeamResponse
// @Failure      404        {object}  models.ErrorResponse  "team not found"
// @Failure      500        {object}  models.ErrorResponse  "internal server error"
// @Router       /team/get [get]
func (c *TeamController) get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()
	teamName := q.Get("team_name")

	team, err := c.teamService.Get(ctx, domain.TeamName(teamName))
	if err != nil {
		if errors.Is(err, domain.ErrTeamNotFound) {
			c.writeError(ctx, w, http.StatusNotFound,
				models.ErrorCodeNotFound,
				"resource not found",
				"team not found",
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
