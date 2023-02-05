package service

import (
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"vpn-wg/internal/model"
	"vpn-wg/internal/store"
	"vpn-wg/internal/util"
)

type WireguardService struct {
	store store.IStore
}

type WireguardServiceInterface interface {
	CreateNew() (model.Peer, error)
}

func NewWireguardService(store store.IStore) *WireguardService {
	return &WireguardService{
		store: store,
	}
}

func (w *WireguardService) CreateNew() (model.Peer, error) {
	peer := model.Peer{}
	server, err := w.store.GetServer()

	if err != nil {
		logrus.Error("Cannot fetch server from database: ", err)
		return peer, err
	}
	allocatedIPs, err := util.GetAllocatedIPs("")

	check, err := util.ValidateIPAllocation(server.Interface.Addresses, allocatedIPs, peer.AllocatedIPs)
	if !check {
		return peer, err
	}

	if util.ValidateAllowedIPs(peer.AllowedIPs) == false {
		logrus.Warnf("Invalid Allowed IPs input from user: %v", peer.AllowedIPs)
		return peer, err
	}

	if util.ValidateExtraAllowedIPs(peer.ExtraAllowedIPs) == false {
		logrus.Warnf("Invalid Extra AllowedIPs input from user: %v", peer.ExtraAllowedIPs)
		return peer, err
	}
	// generate ID
	PeerUuid := uuid.NewV4()
	peer.ID = PeerUuid.String()

	if peer.PublicKey == "" {
		key, err := wgtypes.GeneratePrivateKey()
		if err != nil {
			logrus.Error("Cannot generate wireguard key pair: ", err)
			return peer, err
		}
		peer.PrivateKey = key.String()
		peer.PublicKey = key.PublicKey().String()
	} else {
		_, err := wgtypes.ParseKey(peer.PublicKey)

		if err != nil {
			logrus.Error("Cannot verify wireguard public key: ", err)
			return peer, err
		}
	}

	return peer, nil
}
