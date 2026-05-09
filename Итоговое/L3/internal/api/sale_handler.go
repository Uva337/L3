package api

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"notification-service/internal/model"
	"notification-service/internal/service"

	"github.com/wb-go/wbf/ginext"
)

type SaleHandler struct {
	service *service.SaleService
}

func NewSaleHandler(s *service.SaleService) *SaleHandler {
	return &SaleHandler{service: s}
}

func (h *SaleHandler) CreateItem(c *ginext.Context) {
	var req model.Sale
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "неверный формат данных"})
		return
	}

	if err := h.service.CreateSale(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, req)
}

func (h *SaleHandler) GetItems(c *ginext.Context) {
	sales, err := h.service.GetSales(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	if sales == nil {
		sales = []*model.Sale{}
	}

	// БОНУС: Экспорт в CSV
	if c.Query("format") == "csv" {
		c.Writer.Header().Set("Content-Type", "text/csv")
		c.Writer.Header().Set("Content-Disposition", `attachment;filename="transactions.csv"`)

		writer := csv.NewWriter(c.Writer)
		// Пишем заголовки
		writer.Write([]string{"ID", "Тип", "Сумма", "Категория", "Дата"})

		// Пишем данные
		for _, s := range sales {
			writer.Write([]string{
				s.ID,
				s.Type,
				fmt.Sprintf("%.2f", s.Amount),
				s.Category,
				s.SaleDate.Format("2006-01-02 15:04"),
			})
		}
		writer.Flush()
		return
	}

	c.JSON(http.StatusOK, sales)
}

func (h *SaleHandler) UpdateItem(c *ginext.Context) {
	id := c.Param("id")
	var req model.Sale
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "неверный формат данных"})
		return
	}

	if err := h.service.UpdateSale(c.Request.Context(), id, &req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ginext.H{"status": "успешно обновлено"})
}

func (h *SaleHandler) DeleteItem(c *ginext.Context) {
	id := c.Param("id")
	if err := h.service.DeleteSale(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ginext.H{"status": "успешно удалено"})
}

func (h *SaleHandler) GetAnalytics(c *ginext.Context) {
	fromStr := c.Query("from")
	toStr := c.Query("to")

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		// Если не передали, берем за последний месяц
		from = time.Now().AddDate(0, -1, 0)
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		to = time.Now()
	}

	// Чтобы включить весь день "to", сдвигаем конец на конец дня
	to = to.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	analytics, err := h.service.GetAnalytics(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}
