package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yohnnn/booking_service/internal/cache"
	"github.com/yohnnn/booking_service/internal/event"
	"github.com/yohnnn/booking_service/internal/models"
	"github.com/yohnnn/booking_service/internal/repository"
	"github.com/yohnnn/booking_service/internal/repository/tx"
)

type BookingService struct {
	bookingRepo   repository.BookingRepository
	concertRepo   repository.ConcertRepository
	cacheRepo     cache.ConcertCacheRepository
	manager       *tx.Manager
	eventProducer event.EventProducer
}

func NewBookingService(
	bookingRepo repository.BookingRepository,
	concertRepo repository.ConcertRepository,
	cacheRepo cache.ConcertCacheRepository,
	manager *tx.Manager,
	eventProducer event.EventProducer,
) *BookingService {
	return &BookingService{
		bookingRepo:   bookingRepo,
		concertRepo:   concertRepo,
		cacheRepo:     cacheRepo,
		manager:       manager,
		eventProducer: eventProducer,
	}
}

func (s *BookingService) CreateBooking(ctx context.Context, userID, concertID uuid.UUID, seat int) (*models.Booking, error) {
	var booking *models.Booking

	err := s.manager.WithTx(ctx, func(ctx context.Context) error {

		if err := s.concertRepo.DecrementSeats(ctx, concertID); err != nil {
			return fmt.Errorf("failed to decrement seats (maybe sold out): %w", err)
		}

		booking = &models.Booking{
			UserID:     userID,
			ConcertID:  concertID,
			SeatNumber: seat,
			Status:     models.BookingStatusConfirmed,
		}

		if err := s.bookingRepo.Create(ctx, booking); err != nil {
			return fmt.Errorf("failed to create booking record: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	_ = s.cacheRepo.Delete(ctx)

	go func() {
		evt := event.BookingCreatedEvent{
			BookingID: booking.ID.String(),
			UserID:    booking.UserID.String(),
			ConcertID: booking.ConcertID.String(),
			Seat:      booking.SeatNumber,
		}

		if err := s.eventProducer.SendBookingCreated(context.Background(), evt); err != nil {
			fmt.Printf("Failed to send booking event: %v\n", err)
		}
	}()

	return booking, nil
}

func (s *BookingService) GetUserBookings(ctx context.Context, userID uuid.UUID) ([]models.Booking, error) {
	return s.bookingRepo.GetByUserID(ctx, userID)
}
