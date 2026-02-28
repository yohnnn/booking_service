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

func TestConcertService_GetAll(t *testing.T) {
	type mockBehavior func(repo *mocks.MockConcertRepository, cache *mocks.MockConcertCacheRepository)

	concerts := []models.Concert{
		{
			ID:             uuid.New(),
			Name:           "Rock Festival",
			Place:          "Stadium",
			Date:           time.Now().Add(24 * time.Hour),
			Price:          100.0,
			TotalSeats:     1000,
			AvailableSeats: 500,
		},
		{
			ID:             uuid.New(),
			Name:           "Jazz Night",
			Place:          "Club",
			Date:           time.Now().Add(48 * time.Hour),
			Price:          50.0,
			TotalSeats:     200,
			AvailableSeats: 100,
		},
	}

	tests := []struct {
		name         string
		mockBehavior mockBehavior
		want         []models.Concert
		wantErr      bool
	}{
		{
			name: "success from cache",
			mockBehavior: func(_ *mocks.MockConcertRepository, cache *mocks.MockConcertCacheRepository) {
				cache.EXPECT().
					Get(gomock.Any()).
					Return(concerts, nil)
			},
			want:    concerts,
			wantErr: false,
		},
		{
			name: "success from db (cache miss)",
			mockBehavior: func(repo *mocks.MockConcertRepository, cache *mocks.MockConcertCacheRepository) {
				cache.EXPECT().
					Get(gomock.Any()).
					Return(nil, nil)
				repo.EXPECT().
					GetAll(gomock.Any()).
					Return(concerts, nil)
				cache.EXPECT().
					Set(gomock.Any(), concerts).
					Return(nil)
			},
			want:    concerts,
			wantErr: false,
		},
		{
			name: "success from db (cache error)",
			mockBehavior: func(repo *mocks.MockConcertRepository, cache *mocks.MockConcertCacheRepository) {
				cache.EXPECT().
					Get(gomock.Any()).
					Return(nil, assert.AnError)
				repo.EXPECT().
					GetAll(gomock.Any()).
					Return(concerts, nil)
				cache.EXPECT().
					Set(gomock.Any(), concerts).
					Return(nil)
			},
			want:    concerts,
			wantErr: false,
		},
		{
			name: "repository error",
			mockBehavior: func(repo *mocks.MockConcertRepository, cache *mocks.MockConcertCacheRepository) {
				cache.EXPECT().
					Get(gomock.Any()).
					Return(nil, nil)
				repo.EXPECT().
					GetAll(gomock.Any()).
					Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			concertRepo := mocks.NewMockConcertRepository(ctrl)
			cacheRepo := mocks.NewMockConcertCacheRepository(ctrl)
			tt.mockBehavior(concertRepo, cacheRepo)

			s := NewConcertService(testLogger(), concertRepo, cacheRepo)

			got, err := s.GetAll(context.Background())
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConcertService_GetByID(t *testing.T) {
	concertID := uuid.New()

	concert := &models.Concert{
		ID:             concertID,
		Name:           "Rock Festival",
		Place:          "Stadium",
		Date:           time.Now().Add(24 * time.Hour),
		Price:          100.0,
		TotalSeats:     1000,
		AvailableSeats: 500,
	}

	type mockBehavior func(repo *mocks.MockConcertRepository)

	tests := []struct {
		name         string
		id           uuid.UUID
		mockBehavior mockBehavior
		want         *models.Concert
		wantErr      bool
		wantErrType  error
	}{
		{
			name: "success",
			id:   concertID,
			mockBehavior: func(repo *mocks.MockConcertRepository) {
				repo.EXPECT().
					GetByID(gomock.Any(), concertID).
					Return(concert, nil)
			},
			want:    concert,
			wantErr: false,
		},
		{
			name: "not found",
			id:   uuid.New(),
			mockBehavior: func(repo *mocks.MockConcertRepository) {
				repo.EXPECT().
					GetByID(gomock.Any(), gomock.Any()).
					Return(nil, models.ErrNotFound)
			},
			wantErr:     true,
			wantErrType: models.ErrNotFound,
		},
		{
			name: "repository error",
			id:   concertID,
			mockBehavior: func(repo *mocks.MockConcertRepository) {
				repo.EXPECT().
					GetByID(gomock.Any(), concertID).
					Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			concertRepo := mocks.NewMockConcertRepository(ctrl)
			cacheRepo := mocks.NewMockConcertCacheRepository(ctrl)
			tt.mockBehavior(concertRepo)

			s := NewConcertService(testLogger(), concertRepo, cacheRepo)

			got, err := s.GetByID(context.Background(), tt.id)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrType != nil {
					assert.ErrorIs(t, err, tt.wantErrType)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
