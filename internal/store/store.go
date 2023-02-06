package store

import "vpn-wg/internal/model"

type IStore interface {
	Init() error
	GetServer() (model.Server, error)
	GetPeers(hasQRCode bool) ([]model.PeerData, error)
	SaveCPeer(client model.Peer) error
}
