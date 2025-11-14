package controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"PrService/src/internal/http_api/models"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
)

type baseController struct {
	validate *validator.Validate
	logger   *slog.Logger
}

func newBaseController(v *validator.Validate, logger *slog.Logger) baseController {
	return baseController{
		validate: v,
		logger:   logger,
	}
}

func (bc *baseController) writeJSON(ctx context.Context, w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if payload == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		requestID := middleware.GetReqID(ctx)

		bc.logger.ErrorContext(ctx, "failed to write JSON response",
			"request_id", requestID,
			"err", err,
		)
	}
}

func (bc *baseController) writeError(
	ctx context.Context,
	w http.ResponseWriter,
	status int,
	code models.ErrorCode,
	message string,
	logMsg string,
	err error,
	fields ...any,
) {
	requestID := middleware.GetReqID(ctx)

	logFields := []any{
		"request_id", requestID,
		"error_code", code,
	}
	if err != nil {
		logFields = append(logFields, "err", err)
	}
	logFields = append(logFields, fields...)

	if status >= 500 {
		bc.logger.ErrorContext(ctx, logMsg, logFields...)
	} else {
		bc.logger.InfoContext(ctx, logMsg, logFields...)
	}

	resp := models.CreateErrorResponse(code, message)
	bc.writeJSON(ctx, w, status, resp)
}

func (bc *baseController) decodeAndValidate(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	dst any,
	reqName string,
) bool {
	dec := json.NewDecoder(r.Body)
	requestID := middleware.GetReqID(ctx)

	if err := dec.Decode(dst); err != nil {
		bc.logger.WarnContext(ctx, "failed to decode "+reqName,
			"request_id", requestID,
			"err", err,
		)
		bc.writeError(ctx, w, http.StatusBadRequest,
			models.ErrorCodeDecodeFailed,
			"invalid request body",
			"invalid "+reqName+" body",
			err,
		)
		return false
	}

	if err := bc.validate.StructCtx(ctx, dst); err != nil {
		bc.logger.WarnContext(ctx, "failed to validate "+reqName,
			"request_id", requestID,
			"err", err,
		)
		bc.writeError(ctx, w, http.StatusBadRequest,
			models.ErrorCodeValidationFailed,
			"validation failed",
			"validation failed for "+reqName,
			err,
		)
		return false
	}

	return true
}
