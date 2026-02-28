package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/yohnnn/booking_service/internal/models"
	"github.com/yohnnn/booking_service/internal/service/mocks"
)

func TestBookingService_CreateBooking(t *testing.T) {
	userID := uuid.New()
	concertID := uuid.New()

	type mockBehavior func(
		bookingRepo *mocks.MockBookingRepository,
		concertRepo *mocks.MockConcertRepository,
		cacheRepo *mocks.MockConcertCacheRepository,
		txManager *mocks.MockTxManager,
		producer *mocks.MockEventProducer,
	)

	tests := []struct {
		name         string
		userID       uuid.UUID
		concertID    uuid.UUID
		seat         int
		mockBehavior mockBehavior
		wantErr      bool
		wantErrType  error
		checkResult  func(t *testing.T, booking *models.Booking)
	}{
		{
			name:      "success",
			userID:    userID,
			concertID: concertID,
			seat:      1,
			mockBehavior: func(
				bookingRepo *mocks.MockBookingRepository,
				concertRepo *mocks.MockConcertRepository,
				cacheRepo *mocks.MockConcertCacheRepository,
				txManager *mocks.MockTxManager,
				producer *mocks.MockEventProducer,
			) {
				txManager.EXPECT().
					WithTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})
				concertRepo.EXPECT().
					DecrementSeats(gomock.Any(), concertID).
					Return(nil)
				bookingRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, b *models.Booking) error {
						b.ID = uuid.New()
						b.CreatedAt = time.Now()
						return nil
					})
				cacheRepo.EXPECT().
					Delete(gomock.Any()).
					Return(nil)
				producer.EXPECT().
					SendBookingCreated(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, booking *models.Booking) {
				t.Helper()
				assert.Equal(t, userID, booking.UserID)
				assert.Equal(t, concertID, booking.ConcertID)
				assert.Equal(t, 1, booking.SeatNumber)
				assert.Equal(t, models.BookingStatusConfirmed, booking.Status)
				assert.NotEqual(t, uuid.Nil, booking.ID)
			},
		},
		{
			name:      "no seats available",
			userID:    userID,
			concertID: concertID,
			seat:      1,
			mockBehavior: func(
				_ *mocks.MockBookingRepository,
				concertRepo *mocks.MockConcertRepository,
				_ *mocks.MockConcertCacheRepository,
				txManager *mocks.MockTxManager,
				_ *mocks.MockEventProducer,
			) {
				txManager.EXPECT().
					WithTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})
				concertRepo.EXPECT().
					DecrementSeats(gomock.Any(), concertID).
					Return(models.ErrNoSeats)
			},
			wantErr:     true,
			wantErrType: models.ErrNoSeats,
		},
		{
			name:      "seat already booked",
			userID:    userID,
			concertID: concertID,
			seat:      5,
			mockBehavior: func(
				bookingRepo *mocks.MockBookingRepository,
				concertRepo *mocks.MockConcertRepository,
				_ *mocks.MockConcertCacheRepository,
				txManager *mocks.MockTxManager,
				_ *mocks.MockEventProducer,
			) {
				txManager.EXPECT().
					WithTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})
				concertRepo.EXPECT().
					DecrementSeats(gomock.Any(), concertID).
					Return(nil)
				bookingRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(models.ErrAlreadyExists)
			},
			wantErr:     true,
			wantErrType: models.ErrAlreadyExists,
		},
		{
			name:      "repository error on create",
			userID:    userID,
			concertID: concertID,
			seat:      1,
			mockBehavior: func(
				bookingRepo *mocks.MockBookingRepository,
				concertRepo *mocks.MockConcertRepository,
				_ *mocks.MockConcertCacheRepository,
				txManager *mocks.MockTxManager,
				_ *mocks.MockEventProducer,
			) {
				txManager.EXPECT().
					WithTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})
				concertRepo.EXPECT().
					DecrementSeats(gomock.Any(), concertID).
					Return(nil)
				bookingRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name:      "transaction error",
			userID:    userID,
			concertID: concertID,
			seat:      1,
			mockBehavior: func(
				_ *mocks.MockBookingRepository,
				_ *mocks.MockConcertRepository,
				_ *mocks.MockConcertCacheRepository,
				txManager *mocks.MockTxManager,
				_ *mocks.MockEventProducer,
			) {
				txManager.EXPECT().
					WithTx(gomock.Any(), gomock.Any()).
					Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			bookingRepo := mocks.NewMockBookingRepository(ctrl)
			concertRepo := mocks.NewMockConcertRepository(ctrl)
			cacheRepo := mocks.NewMockConcertCacheRepository(ctrl)
			txManager := mocks.NewMockTxManager(ctrl)
			producer := mocks.NewMockEventProducer(ctrl)

			tt.mockBehavior(bookingRepo, concertRepo, cacheRepo, txManager, producer)

			s := NewBookingService(testLogger(), bookingRepo, concertRepo, cacheRepo, txManager, producer)

			booking, err := s.CreateBooking(context.Background(), tt.userID, tt.concertID, tt.seat)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrType != nil {
					assert.ErrorIs(t, err, tt.wantErrType)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, booking)
			if tt.checkResult != nil {
				tt.checkResult(t, booking)
			}

			s.Wait()
		})
	}
}

func TestBookingService_GetUserBookings(t *testing.T) {
	userID := uuid.New()

	bookings := []models.Booking{
		{
			ID:         uuid.New(),
			UserID:     userID,
			ConcertID:  uuid.New(),
			SeatNumber: 1,
			Status:     models.BookingStatusConfirmed,
			CreatedAt:  time.Now(),
		},
		{
			ID:         uuid.New(),
			UserID:     userID,
			ConcertID:  uuid.New(),
			SeatNumber: 5,
			Status:     models.BookingStatusConfirmed,
			CreatedAt:  time.Now(),
		},
	}

	type mockBehavior func(repo *mocks.MockBookingRepository)

	tests := []struct {
		name         string
		userID       uuid.UUID
		mockBehavior mockBehavior
		want         []models.Booking
		wantErr      bool
	}{
		{
			name:   "success",
			userID: userID,
			mockBehavior: func(repo *mocks.MockBookingRepository) {
				repo.EXPECT().
					GetByUserID(gomock.Any(), userID).
					Return(bookings, nil)
			},
			want:    bookings,
			wantErr: false,
		},
		{
			name:   "empty bookings",
			userID: userID,
			mockBehavior: func(repo *mocks.MockBookingRepository) {
				repo.EXPECT().
					GetByUserID(gomock.Any(), userID).
					Return(nil, nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:   "repository error",
			userID: userID,
			mockBehavior: func(repo *mocks.MockBookingRepository) {
				repo.EXPECT().
					GetByUserID(gomock.Any(), userID).
					Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			bookingRepo := mocks.NewMockBookingRepository(ctrl)
			tt.mockBehavior(bookingRepo)

			s := NewBookingService(
				testLogger(),
				bookingRepo,
				mocks.NewMockConcertRepository(ctrl),
				mocks.NewMockConcertCacheRepository(ctrl),
				mocks.NewMockTxManager(ctrl),
				mocks.NewMockEventProducer(ctrl),
			)

			got, err := s.GetUserBookings(context.Background(), tt.userID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
