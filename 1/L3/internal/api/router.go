package api

import (
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/logger"
)

// SetupRouter создает и настраивает HTTP-сервер
func SetupRouter(handler *Handler, log logger.Logger) *ginext.Engine {
	r := ginext.New("notification-service")

	r.StaticFile("/", "./web/index.html")
	r.Static("/web", "./web")

	r.POST("/notify", handler.Create)
	r.GET("/notify/:id", handler.GetStatus)
	r.DELETE("/notify/:id", handler.Cancel)

	return r
}
