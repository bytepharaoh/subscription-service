package service_test

import (
	"context"
	"testing"
	"time"

	"log/slog"
	"os"

	"github.com/bytepharaoh/subscription-service/internal/domain"
	apperrors "github.com/bytepharaoh/subscription-service/internal/errors"
	"github.com/bytepharaoh/subscription-service/internal/repository"
	"github.com/bytepharaoh/subscription-service/internal/service"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func newLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func newService(t *testing.T) (*service.SubscriptionService, *repository.MockSubscriptionRepository) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mock := repository.NewMockSubscriptionRepository(ctrl)
	svc := service.NewSubscriptionService(mock, newLogger())
	return svc, mock
}

func fixedSub(id uuid.UUID) domain.Subscription {
	return domain.Subscription{
		ID:          id,
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.New(),
		StartDate:   time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestSubscriptionService_Create(t *testing.T) {
	validInput := domain.CreateSubscriptionInput{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.New(),
		StartDate:   "07-2025",
	}

	tests := []struct {
		name      string
		input     domain.CreateSubscriptionInput
		mockSetup func(mock *repository.MockSubscriptionRepository)
		wantErr   bool
		errCode   string
	}{
		{
			name:  "success",
			input: validInput,
			mockSetup: func(mock *repository.MockSubscriptionRepository) {
				mock.EXPECT().
					Create(gomock.Any(), validInput).
					Return(fixedSub(uuid.New()), nil)
			},
			wantErr: false,
		},
		{
			name:  "repository error returns internal error",
			input: validInput,
			mockSetup: func(mock *repository.MockSubscriptionRepository) {
				mock.EXPECT().
					Create(gomock.Any(), validInput).
					Return(domain.Subscription{}, apperrors.ErrInternal)
			},
			wantErr: true,
			errCode: "INTERNAL_ERROR",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, mock := newService(t)
			tc.mockSetup(mock)

			got, err := svc.Create(context.Background(), tc.input)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if appErr, ok := apperrors.IsAppError(err); ok {
					if appErr.Code != tc.errCode {
						t.Errorf("expected error code %q, got %q", tc.errCode, appErr.Code)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ServiceName != tc.input.ServiceName {
				t.Errorf("expected service name %q, got %q", tc.input.ServiceName, got.ServiceName)
			}
		})
	}
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestSubscriptionService_GetByID(t *testing.T) {
	id := uuid.New()
	sub := fixedSub(id)

	tests := []struct {
		name      string
		id        uuid.UUID
		mockSetup func(mock *repository.MockSubscriptionRepository)
		wantErr   bool
		errCode   string
	}{
		{
			name: "success",
			id:   id,
			mockSetup: func(mock *repository.MockSubscriptionRepository) {
				mock.EXPECT().
					GetByID(gomock.Any(), id).
					Return(sub, nil)
			},
			wantErr: false,
		},
		{
			name: "not found",
			id:   id,
			mockSetup: func(mock *repository.MockSubscriptionRepository) {
				mock.EXPECT().
					GetByID(gomock.Any(), id).
					Return(domain.Subscription{}, apperrors.ErrNotFound)
			},
			wantErr: true,
			errCode: "NOT_FOUND",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, mock := newService(t)
			tc.mockSetup(mock)

			got, err := svc.GetByID(context.Background(), tc.id)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if appErr, ok := apperrors.IsAppError(err); ok {
					if appErr.Code != tc.errCode {
						t.Errorf("expected error code %q, got %q", tc.errCode, appErr.Code)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ID != tc.id {
				t.Errorf("expected id %v, got %v", tc.id, got.ID)
			}
		})
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestSubscriptionService_Update(t *testing.T) {
	id := uuid.New()
	sub := fixedSub(id)
	validInput := domain.UpdateSubscriptionInput{
		ServiceName: "Netflix",
		Price:       799,
		StartDate:   "01-2025",
	}

	tests := []struct {
		name      string
		id        uuid.UUID
		input     domain.UpdateSubscriptionInput
		mockSetup func(mock *repository.MockSubscriptionRepository)
		wantErr   bool
		errCode   string
	}{
		{
			name:  "success",
			id:    id,
			input: validInput,
			mockSetup: func(mock *repository.MockSubscriptionRepository) {
				mock.EXPECT().
					GetByID(gomock.Any(), id).
					Return(sub, nil)
				mock.EXPECT().
					Update(gomock.Any(), id, validInput).
					Return(domain.Subscription{
						ID:          id,
						ServiceName: validInput.ServiceName,
						Price:       validInput.Price,
					}, nil)
			},
			wantErr: false,
		},
		{
			name:  "not found on get",
			id:    id,
			input: validInput,
			mockSetup: func(mock *repository.MockSubscriptionRepository) {
				mock.EXPECT().
					GetByID(gomock.Any(), id).
					Return(domain.Subscription{}, apperrors.ErrNotFound)
			},
			wantErr: true,
			errCode: "NOT_FOUND",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, mock := newService(t)
			tc.mockSetup(mock)

			got, err := svc.Update(context.Background(), tc.id, tc.input)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if appErr, ok := apperrors.IsAppError(err); ok {
					if appErr.Code != tc.errCode {
						t.Errorf("expected error code %q, got %q", tc.errCode, appErr.Code)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ServiceName != tc.input.ServiceName {
				t.Errorf("expected service name %q, got %q", tc.input.ServiceName, got.ServiceName)
			}
		})
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestSubscriptionService_Delete(t *testing.T) {
	id := uuid.New()
	sub := fixedSub(id)

	tests := []struct {
		name      string
		id        uuid.UUID
		mockSetup func(mock *repository.MockSubscriptionRepository)
		wantErr   bool
		errCode   string
	}{
		{
			name: "success",
			id:   id,
			mockSetup: func(mock *repository.MockSubscriptionRepository) {
				mock.EXPECT().
					GetByID(gomock.Any(), id).
					Return(sub, nil)
				mock.EXPECT().
					Delete(gomock.Any(), id).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "not found",
			id:   id,
			mockSetup: func(mock *repository.MockSubscriptionRepository) {
				mock.EXPECT().
					GetByID(gomock.Any(), id).
					Return(domain.Subscription{}, apperrors.ErrNotFound)
			},
			wantErr: true,
			errCode: "NOT_FOUND",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, mock := newService(t)
			tc.mockSetup(mock)

			err := svc.Delete(context.Background(), tc.id)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if appErr, ok := apperrors.IsAppError(err); ok {
					if appErr.Code != tc.errCode {
						t.Errorf("expected error code %q, got %q", tc.errCode, appErr.Code)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// ── GetTotalCost ──────────────────────────────────────────────────────────────

func TestSubscriptionService_GetTotalCost(t *testing.T) {
	userID := uuid.New()
	svcName := "Yandex Plus"

	tests := []struct {
		name      string
		input     domain.TotalCostInput
		mockSetup func(mock *repository.MockSubscriptionRepository)
		wantTotal int32
		wantErr   bool
		errCode   string
	}{
		{
			name: "success no filters",
			input: domain.TotalCostInput{
				PeriodStart: "01-2025",
				PeriodEnd:   "06-2025",
			},
			mockSetup: func(mock *repository.MockSubscriptionRepository) {
				mock.EXPECT().
					GetTotalCost(gomock.Any(), gomock.Any()).
					Return(int32(1200), nil)
			},
			wantTotal: 1200,
			wantErr:   false,
		},
		{
			name: "success with user filter",
			input: domain.TotalCostInput{
				UserID:      &userID,
				PeriodStart: "01-2025",
				PeriodEnd:   "06-2025",
			},
			mockSetup: func(mock *repository.MockSubscriptionRepository) {
				mock.EXPECT().
					GetTotalCost(gomock.Any(), gomock.Any()).
					Return(int32(400), nil)
			},
			wantTotal: 400,
			wantErr:   false,
		},
		{
			name: "success with service name filter",
			input: domain.TotalCostInput{
				ServiceName: &svcName,
				PeriodStart: "01-2025",
				PeriodEnd:   "06-2025",
			},
			mockSetup: func(mock *repository.MockSubscriptionRepository) {
				mock.EXPECT().
					GetTotalCost(gomock.Any(), gomock.Any()).
					Return(int32(800), nil)
			},
			wantTotal: 800,
			wantErr:   false,
		},
		{
			name: "invalid period start",
			input: domain.TotalCostInput{
				PeriodStart: "invalid",
				PeriodEnd:   "06-2025",
			},
			mockSetup: func(mock *repository.MockSubscriptionRepository) {},
			wantErr:   true,
			errCode:   "INVALID_DATE",
		},
		{
			name: "period start after period end",
			input: domain.TotalCostInput{
				PeriodStart: "06-2025",
				PeriodEnd:   "01-2025",
			},
			mockSetup: func(mock *repository.MockSubscriptionRepository) {},
			wantErr:   true,
			errCode:   "INVALID_PERIOD",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, mock := newService(t)
			tc.mockSetup(mock)

			got, err := svc.GetTotalCost(context.Background(), tc.input)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if appErr, ok := apperrors.IsAppError(err); ok {
					if appErr.Code != tc.errCode {
						t.Errorf("expected error code %q, got %q", tc.errCode, appErr.Code)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Total != tc.wantTotal {
				t.Errorf("expected total %d, got %d", tc.wantTotal, got.Total)
			}
			if got.Currency != "RUB" {
				t.Errorf("expected currency RUB, got %s", got.Currency)
			}
		})
	}
}
