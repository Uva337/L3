package api

import (
	"context"
	"net/http"
	"time"

	"notification-service/internal/model" // <-- ДОБАВИЛИ ИМПОРТ

	"github.com/wb-go/wbf/ginext"
)

// Теперь интерфейс строго требует возвращать модель уведомления
type Notifier interface {
	CreateNotification(ctx context.Context, message, recipient, channel string, scheduledAt time.Time) (*model.Notification, error)
	GetStatus(ctx context.Context, id string) (*model.Notification, error)
	CancelNotification(ctx context.Context, id string) error
}

// 2. Хендлер теперь зависит от интерфейса, а не от конкретной реализации
type Handler struct {
	service Notifier
}

// 3. Конструктор принимает любой объект, который соответствует интерфейсу Notifier
func NewHandler(s Notifier) *Handler {
	return &Handler{service: s}
}

// Структура того, что мы ожидаем получить от пользователя (Postman)
type createRequest struct {
	Message     string    `json:"message" binding:"required"`
	Recipient   string    `json:"recipient" binding:"required"`
	Channel     string    `json:"channel" binding:"required"` // email или telegram
	ScheduledAt time.Time `json:"scheduled_at" binding:"required"`
}

// Create — ручка (endpoint) для создания уведомления
func (h *Handler) Create(c *ginext.Context) {
	var req createRequest

	// Пытаемся распарсить JSON. Если пользователь прислал кривой запрос — выдаем ошибку 400.
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "Неверный формат запроса: " + err.Error()})
		return
	}

	// Передаем данные в слой бизнес-логики (теперь это вызов через интерфейс!)
	notification, err := h.service.CreateNotification(c.Request.Context(), req.Message, req.Recipient, req.Channel, req.ScheduledAt)
	if err != nil {
		// Ошибка 500, если что-то сломалось в базе
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "Ошибка сохранения: " + err.Error()})
		return
	}

	// Возвращаем успешный ответ (201 Created)
	c.JSON(http.StatusCreated, notification)
}

// GetStatus — ручка (endpoint) для получения статуса
func (h *Handler) GetStatus(c *ginext.Context) {
	// Достаем ID прямо из URL (например, из /notify/123 достанем "123")
	id := c.Param("id")

	// Идем в сервис за данными
	notification, err := h.service.GetStatus(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ginext.H{"error": "Уведомление не найдено"})
		return
	}

	// Отдаем успешный ответ (200 OK)
	c.JSON(http.StatusOK, notification)
}

// Cancel — ручка для отмены уведомления
func (h *Handler) Cancel(c *ginext.Context) {
	id := c.Param("id")

	err := h.service.CancelNotification(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "Не удалось отменить уведомление"})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"message": "Уведомление успешно отменено"})
}
