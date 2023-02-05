package service

import (
	"vpn-wg/internal/model"
	"vpn-wg/internal/store"
)

type WireguardService struct {
	store store.IStore
}

type WireguardServiceInterface interface {
	CreateNew() model.Peer
}

func NewWireguardService(store store.IStore) *WireguardService {
	return &WireguardService{
		store: store,
	}
}

func (w *WireguardService) CreateNew() model.Peer {
	var peer model.Peer
	return peer
}
