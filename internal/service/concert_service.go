package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/yohnnn/booking_service/internal/cache"
	"github.com/yohnnn/booking_service/internal/models"
	"github.com/yohnnn/booking_service/internal/repository"
)

type ConcertService struct {
	concertRepo repository.ConcertRepository
	cacheRepo   cache.ConcertCacheRepository
}

func NewConcertService(concertRepo repository.ConcertRepository, cacheRepo cache.ConcertCacheRepository) *ConcertService {
	return &ConcertService{
		concertRepo: concertRepo,
		cacheRepo:   cacheRepo,
	}
}

func (s *ConcertService) GetAll(ctx context.Context) ([]models.Concert, error) {

	if concerts, err := s.cacheRepo.Get(ctx); err == nil && concerts != nil {
		return concerts, nil
	}

	concerts, err := s.concertRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	_ = s.cacheRepo.Set(ctx, concerts)

	return concerts, nil
}

func (s *ConcertService) GetByID(ctx context.Context, id uuid.UUID) (*models.Concert, error) {
	return s.concertRepo.GetByID(ctx, id)
}
