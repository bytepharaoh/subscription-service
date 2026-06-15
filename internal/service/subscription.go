package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/bytepharaoh/subscription-service/internal/domain"
	apperrors "github.com/bytepharaoh/subscription-service/internal/errors"
	"github.com/bytepharaoh/subscription-service/internal/repository"
	"github.com/google/uuid"
)

type SubscriptionService struct {
	repo   repository.SubscriptionRepository
	logger *slog.Logger
}

func NewSubscriptionService(repo repository.SubscriptionRepository, logger *slog.Logger) *SubscriptionService {
	return &SubscriptionService{
		repo:   repo,
		logger: logger,
	}
}

func (s *SubscriptionService) Create(ctx context.Context, input domain.CreateSubscriptionInput) (domain.Subscription, error) {
	s.logger.Info("creating subscription",
		slog.String("service_name", input.ServiceName),
		slog.String("user_id", input.UserID.String()),
	)

	sub, err := s.repo.Create(ctx, input)
	if err != nil {
		return domain.Subscription{}, err
	}

	s.logger.Info("subscription created", slog.String("id", sub.ID.String()))
	return sub, nil
}

func (s *SubscriptionService) GetByID(ctx context.Context, id uuid.UUID) (domain.Subscription, error) {
	s.logger.Info("fetching subscription", slog.String("id", id.String()))

	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Subscription{}, err
	}

	return sub, nil
}

func (s *SubscriptionService) Update(ctx context.Context, id uuid.UUID, input domain.UpdateSubscriptionInput) (domain.Subscription, error) {
	s.logger.Info("updating subscription", slog.String("id", id.String()))

	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Subscription{}, err
	}

	sub, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return domain.Subscription{}, err
	}

	s.logger.Info("subscription updated", slog.String("id", sub.ID.String()))
	return sub, nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	s.logger.Info("deleting subscription", slog.String("id", id.String()))

	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	s.logger.Info("subscription deleted", slog.String("id", id.String()))
	return nil
}

func (s *SubscriptionService) List(ctx context.Context, input domain.ListSubscriptionsInput) ([]domain.Subscription, error) {
	s.logger.Info("listing subscriptions",
		slog.Int("page", input.Page),
		slog.Int("page_size", input.PageSize),
	)

	return s.repo.List(ctx, input)
}

func (s *SubscriptionService) GetTotalCost(ctx context.Context, input domain.TotalCostInput) (domain.TotalCostResult, error) {
	s.logger.Info("calculating total cost",
		slog.String("period_start", input.PeriodStart),
		slog.String("period_end", input.PeriodEnd),
	)

	// validate dates in service layer before hitting the repo
	periodStart, err := time.Parse(domain.DateLayout, input.PeriodStart)
	if err != nil {
		return domain.TotalCostResult{}, apperrors.New(apperrors.ErrInvalidDate, "period_start: "+input.PeriodStart)
	}

	periodEnd, err := time.Parse(domain.DateLayout, input.PeriodEnd)
	if err != nil {
		return domain.TotalCostResult{}, apperrors.New(apperrors.ErrInvalidDate, "period_end: "+input.PeriodEnd)
	}

	if !periodStart.Before(periodEnd) {
		return domain.TotalCostResult{}, apperrors.ErrInvalidPeriod
	}

	total, err := s.repo.GetTotalCost(ctx, input)
	if err != nil {
		return domain.TotalCostResult{}, err
	}

	return domain.TotalCostResult{
		Total:       total,
		Currency:    "RUB",
		PeriodStart: input.PeriodStart,
		PeriodEnd:   input.PeriodEnd,
	}, nil
}
