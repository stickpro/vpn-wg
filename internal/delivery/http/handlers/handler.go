package handlers

import (
	"github.com/gin-gonic/gin"
	"vpn-wg/internal/service"
)

type Handler struct {
	services *service.Services
}

func NewHandler(services *service.Services) *Handler {
	return &Handler{
		services: services,
	}
}

func (h *Handler) Init(api *gin.RouterGroup) {
	v1 := api.Group("/v1")
	{
		h.initServerRoutes(v1)
		h.initPeerRoutes(v1)
	}
}
