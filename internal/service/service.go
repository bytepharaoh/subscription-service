package service

import (
	"context"

	"github.com/bytepharaoh/subscription-service/internal/domain"
	"github.com/google/uuid"
)

type Subscription interface {
	Create(ctx context.Context, input domain.CreateSubscriptionInput) (domain.Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Subscription, error)
	Update(ctx context.Context, id uuid.UUID, input domain.UpdateSubscriptionInput) (domain.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, input domain.ListSubscriptionsInput) ([]domain.Subscription, error)
	GetTotalCost(ctx context.Context, input domain.TotalCostInput) (domain.TotalCostResult, error)
}
