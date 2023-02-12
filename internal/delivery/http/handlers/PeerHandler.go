package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"vpn-wg/internal/model"
)

func (h *Handler) PeerCreate(c *gin.Context) {
	peerData := model.Peer{Enabled: true}
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
	fmt.Println("[id]", c)
	id := c.Params.ByName("id")
	peer := model.Peer{}

	if err := c.ShouldBindJSON(&peer); err == nil {
		peerData, err := h.services.WireguardService.EditPeer(id, peer)
		if err != nil {
			newResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
		c.JSON(http.StatusOK, peerData)
	} else {
		newResponse(c, http.StatusUnprocessableEntity, err.Error())
	}

}

func (h *Handler) initPeerRoutes(api *gin.RouterGroup) {
	peers := api.Group("/peers")
	{
		peers.POST("", h.PeerCreate)
		peers.PUT("/:id", h.PeerEdit)
	}
}
