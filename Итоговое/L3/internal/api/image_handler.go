package api

import (
	"net/http"
	"os"
	"path/filepath"

	"notification-service/internal/model"
	"notification-service/internal/service"

	"github.com/wb-go/wbf/ginext"
)

type ImageHandler struct {
	service *service.ImageService
}

func NewImageHandler(s *service.ImageService) *ImageHandler {
	return &ImageHandler{service: s}
}

// Upload — POST /upload (Принимает файл от пользователя)
func (h *ImageHandler) Upload(c *ginext.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "Файл не найден в запросе (ключ должен быть 'file')"})
		return
	}

	img, err := h.service.Upload(c.Request.Context(), file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, img)
}

// GetImage — GET /image/:id (Отдает готовую картинку)
func (h *ImageHandler) GetImage(c *ginext.Context) {
	id := c.Param("id")

	img, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ginext.H{"error": "Картинка не найдена в базе"})
		return
	}

	if img.Status != "ready" {
		c.JSON(http.StatusLocked, ginext.H{"error": "Изображение еще находится в обработке"})
		return
	}

	path := filepath.Join("uploads", "processed", img.Filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, ginext.H{"error": "Готовый файл не найден на диске"})
		return
	}

	c.File(path)
}

// Delete — DELETE /image/:id (Удаление)
func (h *ImageHandler) Delete(c *ginext.Context) {
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"status": "успешно удалено"})
}

// GetAll — GET /images (Отдает список всех картинок для UI)
func (h *ImageHandler) GetAll(c *ginext.Context) {
	images, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	if images == nil {
		images = []*model.Image{}
	}

	c.JSON(http.StatusOK, images)
}
