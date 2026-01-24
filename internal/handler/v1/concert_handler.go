package v1

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/yohnnn/booking_service/internal/handler/response"
	"github.com/yohnnn/booking_service/internal/service"
)

type ConcertHandler struct {
	logger  *slog.Logger
	service service.Concert
}

func NewConcertHandler(logger *slog.Logger, service service.Concert) *ConcertHandler {
	return &ConcertHandler{
		logger:  logger,
		service: service,
	}
}

func (h *ConcertHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	concerts, err := h.service.GetAll(r.Context())
	if err != nil {
		h.logger.Error("failed to get concerts", "error", err)
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.ErrCodeInternal, err.Error())
		return
	}

	response.WriteJSONResponse(w, http.StatusOK, concerts)
}

func (h *ConcertHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Warn("invalid concert id", "error", err, "id", idStr)
		response.WriteErrorResponse(w, http.StatusBadRequest, response.ErrCodeInvalidFormat, "invalid concert id")
		return
	}

	concert, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get concert", "error", err)
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.ErrCodeInternal, err.Error())
		return
	}

	response.WriteJSONResponse(w, http.StatusOK, concert)
}
