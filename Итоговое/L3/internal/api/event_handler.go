package api

import (
	"net/http"

	"notification-service/internal/model"
	"notification-service/internal/service"

	"github.com/wb-go/wbf/ginext"
)

type EventHandler struct {
	service *service.EventService
}

func NewEventHandler(s *service.EventService) *EventHandler {
	return &EventHandler{service: s}
}

// CreateEvent — POST /events (Создание мероприятия - для админки)
func (h *EventHandler) CreateEvent(c *ginext.Context) {
	var req model.Event
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "неверный формат запроса"})
		return
	}

	if err := h.service.CreateEvent(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// GetEvents — GET /events (Список всех мероприятий)
func (h *EventHandler) GetEvents(c *ginext.Context) {
	events, err := h.service.GetAllEvents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	if events == nil {
		events = []*model.Event{}
	}
	c.JSON(http.StatusOK, events)
}

// BookSpot — POST /events/:id/book (Бронирование места)
func (h *EventHandler) BookSpot(c *ginext.Context) {
	eventID := c.Param("id")

	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "нужно указать user_id"})
		return
	}

	booking, err := h.service.BookSpot(c.Request.Context(), eventID, req.UserID)
	if err != nil {
		c.JSON(http.StatusConflict, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, booking)
}

// ConfirmBooking — POST /bookings/:id/confirm (Подтверждение/Оплата брони)
func (h *EventHandler) ConfirmBooking(c *ginext.Context) {
	bookingID := c.Param("id")

	if err := h.service.ConfirmBooking(c.Request.Context(), bookingID); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"status": "успешно подтверждено"})
}

// GetEventBookings — GET /events/:id/bookings (Список броней - для админки)
func (h *EventHandler) GetEventBookings(c *ginext.Context) {
	eventID := c.Param("id")

	bookings, err := h.service.GetEventBookings(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	if bookings == nil {
		bookings = []*model.Booking{}
	}
	c.JSON(http.StatusOK, bookings)
}

func (h *EventHandler) GetEvent(c *ginext.Context) {
	id := c.Param("id")
	event, err := h.service.GetEventByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ginext.H{"error": "мероприятие не найдено"})
		return
	}
	c.JSON(http.StatusOK, event)
}
