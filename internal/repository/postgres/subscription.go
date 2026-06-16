package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/bytepharaoh/subscription-service/internal/domain"
	apperrors "github.com/bytepharaoh/subscription-service/internal/errors"
	db "github.com/bytepharaoh/subscription-service/internal/repository/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type SubscriptionRepo struct {
	queries *db.Queries
	logger  *slog.Logger
}

func NewSubscriptionRepo(queries *db.Queries, logger *slog.Logger) *SubscriptionRepo {
	return &SubscriptionRepo{
		queries: queries,
		logger:  logger,
	}
}

func (r *SubscriptionRepo) Create(ctx context.Context, input domain.CreateSubscriptionInput) (domain.Subscription, error) {
	startDate, err := parseMonthYear(input.StartDate)
	if err != nil {
		return domain.Subscription{}, apperrors.New(apperrors.ErrInvalidDate, "start_date: "+input.StartDate)
	}

	params := db.CreateSubscriptionParams{
		ServiceName: input.ServiceName,
		Price:       input.Price,
		UserID:      input.UserID,
		StartDate:   timeToPgDate(startDate),
		EndDate:     pgtype.Date{Valid: false},
	}

	if input.EndDate != nil {
		endDate, err := parseMonthYear(*input.EndDate)
		if err != nil {
			return domain.Subscription{}, apperrors.New(apperrors.ErrInvalidDate, "end_date: "+*input.EndDate)
		}
		params.EndDate = timeToPgDate(endDate)
	}

	row, err := r.queries.CreateSubscription(ctx, params)
	if err != nil {
		r.logger.Error("failed to create subscription", slog.Any("error", err))
		return domain.Subscription{}, fmt.Errorf("create subscription: %w", apperrors.ErrInternal)
	}

	return toDomain(row), nil
}

func (r *SubscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Subscription, error) {
	row, err := r.queries.GetSubscriptionByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Subscription{}, apperrors.ErrNotFound
		}
		r.logger.Error("failed to get subscription", slog.String("id", id.String()), slog.Any("error", err))
		return domain.Subscription{}, fmt.Errorf("get subscription: %w", apperrors.ErrInternal)
	}

	return toDomain(row), nil
}

func (r *SubscriptionRepo) Update(ctx context.Context, id uuid.UUID, input domain.UpdateSubscriptionInput) (domain.Subscription, error) {
	startDate, err := parseMonthYear(input.StartDate)
	if err != nil {
		return domain.Subscription{}, apperrors.New(apperrors.ErrInvalidDate, "start_date: "+input.StartDate)
	}

	params := db.UpdateSubscriptionParams{
		ID:          id,
		ServiceName: input.ServiceName,
		Price:       input.Price,
		StartDate:   timeToPgDate(startDate),
		EndDate:     pgtype.Date{Valid: false},
	}

	if input.EndDate != nil {
		endDate, err := parseMonthYear(*input.EndDate)
		if err != nil {
			return domain.Subscription{}, apperrors.New(apperrors.ErrInvalidDate, "end_date: "+*input.EndDate)
		}
		params.EndDate = timeToPgDate(endDate)
	}

	row, err := r.queries.UpdateSubscription(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Subscription{}, apperrors.ErrNotFound
		}
		r.logger.Error("failed to update subscription", slog.String("id", id.String()), slog.Any("error", err))
		return domain.Subscription{}, fmt.Errorf("update subscription: %w", apperrors.ErrInternal)
	}

	return toDomain(row), nil
}

func (r *SubscriptionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteSubscription(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		r.logger.Error("failed to delete subscription", slog.String("id", id.String()), slog.Any("error", err))
		return fmt.Errorf("delete subscription: %w", apperrors.ErrInternal)
	}

	return nil
}

func (r *SubscriptionRepo) List(ctx context.Context, input domain.ListSubscriptionsInput) ([]domain.Subscription, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 {
		input.PageSize = 20
	}

	params := db.ListSubscriptionsParams{
		OffsetCount: int32((input.Page - 1) * input.PageSize),
		LimitCount:  int32(input.PageSize),
	}

	if input.UserID != nil {
		uid, err := uuid.Parse(*input.UserID)
		if err != nil {
			return nil, apperrors.New(apperrors.ErrInvalidInput, "invalid user_id: "+*input.UserID)
		}
		params.UserID = pgtype.UUID{Bytes: uid, Valid: true}
	}

	if input.ServiceName != nil {
		params.ServiceName = pgtype.Text{String: *input.ServiceName, Valid: true}
	}

	rows, err := r.queries.ListSubscriptions(ctx, params)
	if err != nil {
		r.logger.Error("failed to list subscriptions", slog.Any("error", err))
		return nil, fmt.Errorf("list subscriptions: %w", apperrors.ErrInternal)
	}

	result := make([]domain.Subscription, 0, len(rows))
	for _, row := range rows {
		result = append(result, toDomain(row))
	}

	return result, nil
}

func (r *SubscriptionRepo) GetTotalCost(ctx context.Context, input domain.TotalCostInput) (int32, error) {
	periodStart, _ := parseMonthYear(input.PeriodStart)
	periodEnd, _ := parseMonthYear(input.PeriodEnd)

	params := db.GetTotalCostParams{
		PeriodStart: timeToPgDate(periodStart),
		PeriodEnd:   timeToPgDate(periodEnd),
	}

	if input.UserID != nil {
		uid, err := uuid.Parse(*input.UserID)
		if err != nil {
			return 0, apperrors.New(apperrors.ErrInvalidInput, "invalid user_id: "+*input.UserID)
		}
		params.UserID = pgtype.UUID{Bytes: uid, Valid: true}
	}

	if input.ServiceName != nil {
		params.ServiceName = pgtype.Text{String: *input.ServiceName, Valid: true}
	}

	total, err := r.queries.GetTotalCost(ctx, params)
	if err != nil {
		r.logger.Error("failed to get total cost", slog.Any("error", err))
		return 0, fmt.Errorf("get total cost: %w", apperrors.ErrInternal)
	}

	return total, nil
}

//  helpers

func parseMonthYear(s string) (time.Time, error) {
	t, err := time.Parse(domain.DateLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q: %w", s, err)
	}
	return t, nil
}

func timeToPgDate(t time.Time) pgtype.Date {
	return pgtype.Date{
		Time:  t,
		Valid: true,
	}
}

func toDomain(row db.Subscription) domain.Subscription {
	sub := domain.Subscription{
		ID:          row.ID,
		ServiceName: row.ServiceName,
		Price:       row.Price,
		UserID:      row.UserID,
		StartDate:   row.StartDate.Time,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}

	if row.EndDate.Valid {
		t := row.EndDate.Time
		sub.EndDate = &t
	}

	return sub
}
