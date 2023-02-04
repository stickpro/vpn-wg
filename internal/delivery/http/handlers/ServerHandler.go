package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (h *Handler) ServerInfo(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func (h *Handler) initServerRoutes(api *gin.RouterGroup) {
	users := api.Group("/server")
	{
		users.GET("", h.ServerInfo)
	}
}
