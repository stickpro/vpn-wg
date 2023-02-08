package store

import "vpn-wg/internal/model"

type IStore interface {
	Init() error
	GetServer() (model.Server, error)
	GetPeers(hasQRCode bool) ([]model.PeerData, error)
	SavePeer(client model.Peer) error
	GetPeerByID(peerID string, qrCode model.QRCodeSettings) (model.PeerData, error)
}
