package service

import "vpn-wg/internal/store"

type Services struct {
	WireguardService WireguardServiceInterface
}

func NewServices(store store.IStore) *Services {
	wireguardService := NewWireguardService(store)

	return &Services{
		WireguardService: wireguardService,
	}
}
