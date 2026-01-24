package v1

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/yohnnn/booking_service/internal/dto"
	"github.com/yohnnn/booking_service/internal/handler/response"
	"github.com/yohnnn/booking_service/internal/middleware"
	"github.com/yohnnn/booking_service/internal/service"
)

type BookingHandler struct {
	logger    *slog.Logger
	validator *validator.Validate
	service   service.Booking
}

func NewBookingHandler(logger *slog.Logger, validator *validator.Validate, service service.Booking) *BookingHandler {
	return &BookingHandler{
		logger:    logger,
		validator: validator,
		service:   service,
	}
}

func (h *BookingHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Error("user id not found in context")
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.ErrCodeInternal, "internal error")
		return
	}

	var input dto.CreateBookingRequest
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

	concertID, err := uuid.Parse(input.ConcertID)
	if err != nil {
		h.logger.Warn("invalid concert id", "error", err)
		response.WriteErrorResponse(w, http.StatusBadRequest, response.ErrCodeInvalidFormat, "invalid concert id")
		return
	}

	booking, err := h.service.CreateBooking(r.Context(), userID, concertID, input.Seat)
	if err != nil {
		h.logger.Error("failed to create booking", "error", err)
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.ErrCodeInternal, err.Error())
		return
	}

	response.WriteJSONResponse(w, http.StatusCreated, booking)
}
func (h *BookingHandler) GetUserBookings(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Error("user id not found in context")
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.ErrCodeInternal, "internal error")
		return
	}

	bookings, err := h.service.GetUserBookings(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get user bookings", "error", err)
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.ErrCodeInternal, err.Error())
		return
	}

	response.WriteJSONResponse(w, http.StatusOK, bookings)
}
