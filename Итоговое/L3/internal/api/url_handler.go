package api

import (
	"net/http"

	"notification-service/internal/service"

	"github.com/wb-go/wbf/ginext"
)

type URLHandler struct {
	service *service.URLShortenerService
}

func NewURLHandler(s *service.URLShortenerService) *URLHandler {
	return &URLHandler{service: s}
}

type shortenRequest struct {
	OriginalURL string `json:"original_url" binding:"required"`
	CustomAlias string `json:"custom_alias"`
}


func (h *URLHandler) Shorten(c *ginext.Context) {
	var req shortenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "Неверный формат запроса: " + err.Error()})
		return
	}

	url, err := h.service.ShortenURL(c.Request.Context(), req.OriginalURL, req.CustomAlias)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "Ошибка сохранения (возможно, имя уже занято): " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ginext.H{
		"id":           url.ID,
		"short_code":   url.ShortCode,
		"short_url":    "http://localhost:8080/s/" + url.ShortCode,
		"original_url": url.OriginalURL,
	})
}

func (h *URLHandler) Redirect(c *ginext.Context) {
	shortCode := c.Param("code")

	userAgent := c.Request.UserAgent()

	originalURL, err := h.service.ProcessRedirect(c.Request.Context(), shortCode, userAgent)
	if err != nil {
		c.JSON(http.StatusNotFound, ginext.H{"error": "Ссылка не найдена или удалена"})
		return
	}

	c.Redirect(http.StatusFound, originalURL)
}

func (h *URLHandler) Analytics(c *ginext.Context) {
	shortCode := c.Param("code")

	stats, err := h.service.GetAnalytics(c.Request.Context(), shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, ginext.H{"error": "Ссылка не найдена"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
