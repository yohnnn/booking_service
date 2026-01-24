package v1

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/yohnnn/booking_service/internal/dto"
	"github.com/yohnnn/booking_service/internal/handler/response"
	"github.com/yohnnn/booking_service/internal/service"
)

type AuthHandler struct {
	logger    *slog.Logger
	validator *validator.Validate
	service   service.Auth
}

func NewAuthHandler(logger *slog.Logger, validator *validator.Validate, service service.Auth) *AuthHandler {
	return &AuthHandler{
		logger:    logger,
		validator: validator,
		service:   service,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Warn("failed to decode request body", "error", err)
		response.WriteErrorResponse(w, http.StatusBadRequest, response.ErrCodeInvalidFormat, "invalid input body")
		return
	}

	if err := h.validator.Struct(input); err != nil {
		h.logger.Warn("validation failed", "error", err)
		response.WriteErrorResponse(w, http.StatusBadRequest, response.ErrCodeValidationFailed, err.Error())
		return
	}

	user, err := h.service.Register(r.Context(), input.Email, input.Password)
	if err != nil {
		h.logger.Error("failed to register user", "error", err)
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.ErrCodeInternal, err.Error())
		return
	}

	response.WriteJSONResponse(w, http.StatusCreated, user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Warn("failed to decode request body", "error", err)
		response.WriteErrorResponse(w, http.StatusBadRequest, response.ErrCodeInvalidFormat, "invalid input body")
		return
	}

	if err := h.validator.Struct(input); err != nil {
		h.logger.Warn("validation failed", "error", err)
		response.WriteErrorResponse(w, http.StatusBadRequest, response.ErrCodeValidationFailed, err.Error())
		return
	}

	token, err := h.service.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		h.logger.Error("failed to login", "error", err)
		response.WriteErrorResponse(w, http.StatusUnauthorized, response.ErrCodeUnauthorized, err.Error())
		return
	}

	response.WriteJSONResponse(w, http.StatusOK, map[string]string{
		"token": token,
	})
}
