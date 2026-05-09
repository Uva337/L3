package api

import (
	"net/http"
	"strconv"

	"notification-service/internal/model"
	"notification-service/internal/service"

	"github.com/wb-go/wbf/ginext"
)

type CommentHandler struct {
	service *service.CommentService
}

func NewCommentHandler(s *service.CommentService) *CommentHandler {
	return &CommentHandler{service: s}
}

type createCommentReq struct {
	ParentID *string `json:"parent_id"`
	Author   string  `json:"author"`
	Text     string  `json:"text" binding:"required"`
}

func (h *CommentHandler) Create(c *ginext.Context) {
	var req createCommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "Неверный формат запроса: " + err.Error()})
		return
	}

	comment, err := h.service.Create(c.Request.Context(), req.ParentID, req.Author, req.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "Ошибка сохранения в БД: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

func (h *CommentHandler) Get(c *ginext.Context) {
	parentID := c.Query("parent")
	searchQuery := c.Query("search")
	ctx := c.Request.Context()

	if searchQuery != "" {
		comments, err := h.service.Search(ctx, searchQuery)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
			return
		}
		if comments == nil {
			comments = []*model.Comment{}
		}
		c.JSON(http.StatusOK, comments)
		return
	}

	if parentID != "" {
		tree, err := h.service.GetTree(ctx, parentID)
		if err != nil || tree == nil {
			c.JSON(http.StatusNotFound, ginext.H{"error": "Комментарий не найден"})
			return
		}

		c.JSON(http.StatusOK, []*model.Comment{tree})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	roots, err := h.service.GetRoots(ctx, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	if roots == nil {
		roots = []*model.Comment{}
	}
	c.JSON(http.StatusOK, roots)
}

func (h *CommentHandler) Delete(c *ginext.Context) {
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "Ошибка при удалении"})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"status": "успешно удалено"})
}
