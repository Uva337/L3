package api

import (
	"context"
	"net/http"
	"time"

	"notification-service/internal/model" // <-- ДОБАВИЛИ ИМПОРТ

	"github.com/wb-go/wbf/ginext"
)

type Notifier interface {
	CreateNotification(ctx context.Context, message, recipient, channel string, scheduledAt time.Time) (*model.Notification, error)
	GetStatus(ctx context.Context, id string) (*model.Notification, error)
	CancelNotification(ctx context.Context, id string) error
}

type Handler struct {
	service Notifier
}

func NewHandler(s Notifier) *Handler {
	return &Handler{service: s}
}

type createRequest struct {
	Message     string    `json:"message" binding:"required"`
	Recipient   string    `json:"recipient" binding:"required"`
	Channel     string    `json:"channel" binding:"required"` // email или telegram
	ScheduledAt time.Time `json:"scheduled_at" binding:"required"`
}

// Create ручка для создания уведомления
func (h *Handler) Create(c *ginext.Context) {
	var req createRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "Неверный формат запроса: " + err.Error()})
		return
	}

	notification, err := h.service.CreateNotification(c.Request.Context(), req.Message, req.Recipient, req.Channel, req.ScheduledAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "Ошибка сохранения: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, notification)
}

// GetStatus ручка для получения статуса
func (h *Handler) GetStatus(c *ginext.Context) {
	id := c.Param("id")

	notification, err := h.service.GetStatus(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ginext.H{"error": "Уведомление не найдено"})
		return
	}

	c.JSON(http.StatusOK, notification)
}

// Cancel ручка для отмены уведомления
func (h *Handler) Cancel(c *ginext.Context) {
	id := c.Param("id")

	err := h.service.CancelNotification(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "Не удалось отменить уведомление"})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"message": "Уведомление успешно отменено"})
}
