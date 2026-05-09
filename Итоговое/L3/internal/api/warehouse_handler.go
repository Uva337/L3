package api

import (
	"net/http"

	"notification-service/internal/model"
	"notification-service/internal/service"

	"github.com/wb-go/wbf/ginext"
)

type WarehouseHandler struct {
	service *service.WarehouseService
}

func NewWarehouseHandler(s *service.WarehouseService) *WarehouseHandler {
	return &WarehouseHandler{service: s}
}

// Вспомогательная функция: достаем имя пользователя из контекста (положили в AuthMiddleware)
func getUsername(c *ginext.Context) string {
	username, exists := c.Get("username")
	if !exists {
		return "system"
	}
	return username.(string)
}

func (h *WarehouseHandler) CreateItem(c *ginext.Context) {
	var req model.Item
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "неверный формат данных"})
		return
	}

	username := getUsername(c)
	if err := h.service.CreateItem(c.Request.Context(), &req, username); err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, req)
}

func (h *WarehouseHandler) GetItems(c *ginext.Context) {
	items, err := h.service.GetItems(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}
	if items == nil {
		items = []*model.Item{}
	}
	c.JSON(http.StatusOK, items)
}

func (h *WarehouseHandler) UpdateItem(c *ginext.Context) {
	id := c.Param("id")
	var req model.Item
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "неверный формат данных"})
		return
	}

	username := getUsername(c)
	if err := h.service.UpdateItem(c.Request.Context(), id, &req, username); err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ginext.H{"status": "обновлено"})
}

func (h *WarehouseHandler) DeleteItem(c *ginext.Context) {
	id := c.Param("id")
	username := getUsername(c)

	if err := h.service.DeleteItem(c.Request.Context(), id, username); err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ginext.H{"status": "удалено"})
}

func (h *WarehouseHandler) GetItemHistory(c *ginext.Context) {
	id := c.Param("id")
	history, err := h.service.GetItemHistory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}
	if history == nil {
		history = []*model.ItemHistory{}
	}
	c.JSON(http.StatusOK, history)
}
