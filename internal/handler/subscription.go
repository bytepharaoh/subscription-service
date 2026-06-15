package handler

import (
	"net/http"

	"github.com/bytepharaoh/subscription-service/internal/domain"
	apperrors "github.com/bytepharaoh/subscription-service/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Create godoc
// @Summary      Create a subscription
// @Description  Create a new subscription record
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        input  body      domain.CreateSubscriptionInput  true  "Subscription input"
// @Success      201    {object}  domain.Subscription
// @Failure      400    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /subscriptions [post]
func (h *Handler) Create(c *gin.Context) {
	var input domain.CreateSubscriptionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.respondError(c, apperrors.New(apperrors.ErrInvalidInput, err.Error()))
		return
	}

	sub, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		h.respondError(c, err)
		return
	}

	h.respondCreated(c, sub)
}

// GetByID godoc
// @Summary      Get a subscription
// @Description  Get a subscription by ID
// @Tags         subscriptions
// @Produce      json
// @Param        id   path      string  true  "Subscription ID"
// @Success      200  {object}  domain.Subscription
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /subscriptions/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		h.respondError(c, err)
		return
	}

	sub, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		h.respondError(c, err)
		return
	}

	h.respondOK(c, sub)
}

// Update godoc
// @Summary      Update a subscription
// @Description  Update an existing subscription by ID
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id     path      string                          true  "Subscription ID"
// @Param        input  body      domain.UpdateSubscriptionInput  true  "Subscription update input"
// @Success      200    {object}  domain.Subscription
// @Failure      400    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /subscriptions/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		h.respondError(c, err)
		return
	}

	var input domain.UpdateSubscriptionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.respondError(c, apperrors.New(apperrors.ErrInvalidInput, err.Error()))
		return
	}

	sub, err := h.service.Update(c.Request.Context(), id, input)
	if err != nil {
		h.respondError(c, err)
		return
	}

	h.respondOK(c, sub)
}

// Delete godoc
// @Summary      Delete a subscription
// @Description  Delete a subscription by ID
// @Tags         subscriptions
// @Produce      json
// @Param        id   path      string  true  "Subscription ID"
// @Success      204  "No Content"
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /subscriptions/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		h.respondError(c, err)
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		h.respondError(c, err)
		return
	}

	h.respondNoContent(c)
}

// List godoc
// @Summary      List subscriptions
// @Description  List subscriptions with optional filters and pagination
// @Tags         subscriptions
// @Produce      json
// @Param        user_id       query     string  false  "Filter by user ID"
// @Param        service_name  query     string  false  "Filter by service name"
// @Param        page          query     int     false  "Page number (default: 1)"
// @Param        page_size     query     int     false  "Page size (default: 20, max: 100)"
// @Success      200           {array}   domain.Subscription
// @Failure      400           {object}  map[string]string
// @Failure      500           {object}  map[string]string
// @Router       /subscriptions [get]
func (h *Handler) List(c *gin.Context) {
	var input domain.ListSubscriptionsInput
	if err := c.ShouldBindQuery(&input); err != nil {
		h.respondError(c, apperrors.New(apperrors.ErrInvalidInput, err.Error()))
		return
	}

	subs, err := h.service.List(c.Request.Context(), input)
	if err != nil {
		h.respondError(c, err)
		return
	}

	h.respondOK(c, subs)
}

// GetTotalCost godoc
// @Summary      Get total subscription cost
// @Description  Calculate total cost of subscriptions for a period with optional filters
// @Tags         subscriptions
// @Produce      json
// @Param        period_start  query     string  true   "Period start in MM-YYYY format"
// @Param        period_end    query     string  true   "Period end in MM-YYYY format"
// @Param        user_id       query     string  false  "Filter by user ID"
// @Param        service_name  query     string  false  "Filter by service name"
// @Success      200           {object}  domain.TotalCostResult
// @Failure      400           {object}  map[string]string
// @Failure      500           {object}  map[string]string
// @Router       /subscriptions/total-cost [get]
func (h *Handler) GetTotalCost(c *gin.Context) {
	var input domain.TotalCostInput
	if err := c.ShouldBindQuery(&input); err != nil {
		h.respondError(c, apperrors.New(apperrors.ErrInvalidInput, err.Error()))
		return
	}

	result, err := h.service.GetTotalCost(c.Request.Context(), input)
	if err != nil {
		h.respondError(c, err)
		return
	}

	h.respondOK(c, result)
}

//  helpers

func parseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	raw := c.Param(param)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.UUID{}, apperrors.New(apperrors.ErrInvalidInput, "invalid UUID: "+raw)
	}
	return id, nil
}

// Health godoc
// @Summary      Health check
// @Description  Returns service health status
// @Tags         system
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
