package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"

	"github.com/yohnnn/booking_service/internal/cache"
	"github.com/yohnnn/booking_service/internal/event"
	"github.com/yohnnn/booking_service/internal/models"
	"github.com/yohnnn/booking_service/internal/repository"
)

type BookingService struct {
	logger        *slog.Logger
	bookingRepo   repository.BookingRepository
	concertRepo   repository.ConcertRepository
	cacheRepo     cache.ConcertCacheRepository
	manager       TxManager
	eventProducer event.EventProducer
	wg            sync.WaitGroup
}

func NewBookingService(
	logger *slog.Logger,
	bookingRepo repository.BookingRepository,
	concertRepo repository.ConcertRepository,
	cacheRepo cache.ConcertCacheRepository,
	manager TxManager,
	eventProducer event.EventProducer,
) *BookingService {
	return &BookingService{
		logger:        logger,
		bookingRepo:   bookingRepo,
		concertRepo:   concertRepo,
		cacheRepo:     cacheRepo,
		manager:       manager,
		eventProducer: eventProducer,
	}
}

func (s *BookingService) CreateBooking(
	ctx context.Context,
	userID, concertID uuid.UUID,
	seat int,
) (*models.Booking, error) {
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

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		evt := event.BookingCreatedEvent{
			BookingID: booking.ID.String(),
			UserID:    booking.UserID.String(),
			ConcertID: booking.ConcertID.String(),
			Seat:      booking.SeatNumber,
		}

		if err := s.eventProducer.SendBookingCreated(context.Background(), evt); err != nil {
			s.logger.Error("failed to send booking event", "error", err)
		}
	}()

	return booking, nil
}

func (s *BookingService) Wait() {
	s.wg.Wait()
}

func (s *BookingService) GetUserBookings(ctx context.Context, userID uuid.UUID) ([]models.Booking, error) {
	return s.bookingRepo.GetByUserID(ctx, userID)
}
