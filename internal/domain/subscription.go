package domain

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID  `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int32      `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateSubscriptionInput struct {
	ServiceName string    `json:"service_name" binding:"required,min=1,max=255"`
	Price       int32     `json:"price"        binding:"required,gt=0"`
	UserID      uuid.UUID `json:"user_id"      binding:"required"`
	StartDate   string    `json:"start_date"   binding:"required"`
	EndDate     *string   `json:"end_date"`
}

type UpdateSubscriptionInput struct {
	ServiceName string  `json:"service_name" binding:"required,min=1,max=255"`
	Price       int32   `json:"price"        binding:"required,gt=0"`
	StartDate   string  `json:"start_date"   binding:"required"`
	EndDate     *string `json:"end_date"`
}

type ListSubscriptionsInput struct {
	UserID      *uuid.UUID `form:"user_id"`
	ServiceName *string    `form:"service_name"`
	Page        int        `form:"page"    binding:"omitempty,min=1"`
	PageSize    int        `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type TotalCostInput struct {
	UserID      *uuid.UUID `form:"user_id"`
	ServiceName *string    `form:"service_name"`
	PeriodStart string     `form:"period_start" binding:"required"`
	PeriodEnd   string     `form:"period_end"   binding:"required"`
}

type TotalCostResult struct {
	Total       int32  `json:"total"`
	Currency    string `json:"currency"`
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
}

const DateLayout = "01-2006"
