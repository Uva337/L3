package api

import (
	"notification-service/internal/middleware"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/logger"
)

func SetupRouter(
	handler *Handler,
	urlHandler *URLHandler,
	commentHandler *CommentHandler,
	imageHandler *ImageHandler,
	eventHandler *EventHandler,
	saleHandler *SaleHandler,
	warehouseHandler *WarehouseHandler,
	log logger.Logger,
) *ginext.Engine {
	r := ginext.New("notification-service")

	//  статика (UI)
	r.StaticFile("/", "./web/index.html")
	r.StaticFile("/shortener", "./web/shortener.html")
	r.StaticFile("/comments-ui", "./web/comments.html")
	r.StaticFile("/gallery", "./web/gallery.html")
	r.StaticFile("/booking", "./web/booking.html")
	r.StaticFile("/admin-booking", "./web/admin_booking.html")
	r.StaticFile("/sales", "./web/sales.html")
	r.StaticFile("/warehouse", "./web/warehouse.html")

	r.Static("/web", "./web")

	// Ручки сервиса уведомлений
	r.POST("/notify", handler.Create)
	r.GET("/notify/:id", handler.GetStatus)
	r.DELETE("/notify/:id", handler.Cancel)

	// Pучки сервиса коротких ссылок
	r.POST("/shorten", urlHandler.Shorten)
	r.GET("/s/:code", urlHandler.Redirect)
	r.GET("/analytics/:code", urlHandler.Analytics)

	// Ручки сервиса комментариев
	r.POST("/comments", commentHandler.Create)
	r.GET("/comments", commentHandler.Get)
	r.DELETE("/comments/:id", commentHandler.Delete)

	//  Ручки сервиса изображений
	r.POST("/upload", imageHandler.Upload)
	r.GET("/images", imageHandler.GetAll)
	r.GET("/image/:id", imageHandler.GetImage)
	r.GET("/events/:id", eventHandler.GetEvent)
	r.DELETE("/image/:id", imageHandler.Delete)

	r.POST("/events", eventHandler.CreateEvent)
	r.GET("/events", eventHandler.GetEvents)
	r.POST("/events/:id/book", eventHandler.BookSpot)
	r.GET("/events/:id/bookings", eventHandler.GetEventBookings)
	r.POST("/bookings/:id/confirm", eventHandler.ConfirmBooking)

	// Ручки сервиса Аналитики и Трекера
	r.POST("/items", saleHandler.CreateItem)
	r.GET("/items", saleHandler.GetItems)
	r.PUT("/items/:id", saleHandler.UpdateItem)
	r.DELETE("/items/:id", saleHandler.DeleteItem)
	r.GET("/analytics", saleHandler.GetAnalytics)

	// Смотреть могут все три роли
	r.GET("/api/warehouse/items", middleware.AuthMiddleware("admin", "manager", "viewer"), warehouseHandler.GetItems)
	r.GET("/api/warehouse/items/:id/history", middleware.AuthMiddleware("admin", "manager", "viewer"), warehouseHandler.GetItemHistory)

	// Создавать и изменять могут только админ и менеджер
	r.POST("/api/warehouse/items", middleware.AuthMiddleware("admin", "manager"), warehouseHandler.CreateItem)
	r.PUT("/api/warehouse/items/:id", middleware.AuthMiddleware("admin", "manager"), warehouseHandler.UpdateItem)
	
	// Удалять имеет право ТОЛЬКО админ
	r.DELETE("/api/warehouse/items/:id", middleware.AuthMiddleware("admin"), warehouseHandler.DeleteItem)

	return r
}
