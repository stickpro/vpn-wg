package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"vpn-wg/internal/model"
)

func (h *Handler) PeerCreate(c *gin.Context) {
	peerData := model.Peer{}
	if err := c.ShouldBindJSON(&peerData); err == nil {
		peer, err := h.services.WireguardService.CreateNew(peerData)
		if err != nil {
			newResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
		c.JSON(http.StatusOK, peer)
	} else {
		newResponse(c, http.StatusUnprocessableEntity, err.Error())
	}
}

func (h *Handler) PeerEdit(c *gin.Context) {
	id := c.Params.ByName("id")
	peer := model.Peer{}

	if err := c.ShouldBindJSON(&peer); err == nil {
		peerData, _ := h.services.WireguardService.EditPeer(id, peer)
	}

	c.JSON(http.StatusOK, peerData)
}

func (h *Handler) initPeerRoutes(api *gin.RouterGroup) {
	peers := api.Group("/peers")
	{
		peers.POST("", h.PeerCreate)
		peers.GET("/:id", h.PeerEdit)
	}
}
