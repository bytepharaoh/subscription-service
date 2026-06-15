package handler

import (
	"net/http"

	apperrors "github.com/bytepharaoh/subscription-service/internal/errors"
	"github.com/bytepharaoh/subscription-service/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service service.Subscription
}

func New(svc service.Subscription) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) respondOK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}

func (h *Handler) respondCreated(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, data)
}

func (h *Handler) respondNoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func (h *Handler) respondError(c *gin.Context, err error) {
	if appErr, ok := apperrors.IsAppError(err); ok {
		c.JSON(appErr.StatusCode, gin.H{
			"code":    appErr.Code,
			"message": appErr.Message,
		})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    "INTERNAL_ERROR",
		"message": "internal server error",
	})
}
