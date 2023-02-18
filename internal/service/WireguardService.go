package service

import (
	"fmt"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"os"
	"strings"
	"text/template"
	"time"
	"vpn-wg/internal/model"
	"vpn-wg/internal/store"
	"vpn-wg/internal/util"
)

type WireguardService struct {
	store store.IStore
}

type WireguardServiceInterface interface {
	CreateNew(peer model.Peer) (model.Peer, error)
	EditPeer(id string, peerValue model.Peer) (model.PeerData, error)
	DeletePeer(id string) error
	applyConfig() error
}

func NewWireguardService(store store.IStore) *WireguardService {
	return &WireguardService{
		store: store,
	}
}

// TODO refactoring method to small function and add text message for error
func (w *WireguardService) CreateNew(peer model.Peer) (model.Peer, error) {
	logrus.Info("[PeersData]", peer)
	server, err := w.store.GetServer()

	if err != nil {
		logrus.Error("Cannot fetch server from database: ", err)
		return peer, err
	}
	allocatedIPs, err := util.GetAllocatedIPs("")
	suggestedIPs := make([]string, 0)

	for _, cidr := range server.Interface.Addresses {
		ip, err := util.GetAvailableIP(cidr, allocatedIPs)
		if err != nil {
			logrus.Error("Failed to get available ip from a CIDR: ", err)
			return peer, err
		}
		if strings.Contains(ip, ":") {
			suggestedIPs = append(suggestedIPs, fmt.Sprintf("%s/128", ip))
		} else {
			suggestedIPs = append(suggestedIPs, fmt.Sprintf("%s/32", ip))
		}
	}

	peer.AllocatedIPs = suggestedIPs
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
		// check for duplicates
		peers, err := w.store.GetPeers(false)
		if err != nil {
			logrus.Error("Cannot get clients for duplicate check")
			return peer, err
		}
		for _, other := range peers {
			if other.Peer.PublicKey == peer.PublicKey {
				logrus.Error("Duplicate Public Key")
				return peer, err
			}
		}
	}

	if peer.PresharedKey == "" {
		presharedKey, err := wgtypes.GenerateKey()
		if err != nil {
			logrus.Error("Cannot generated preshared key: ", err)
			return peer, err
		}
		peer.PresharedKey = presharedKey.String()
	} else if peer.PresharedKey == "-" {
		peer.PresharedKey = ""
		logrus.Infof("skipped PresharedKey generation for user: %v", peer.Name)
	} else {
		_, err := wgtypes.ParseKey(peer.PresharedKey)
		if err != nil {
			logrus.Error("Cannot verify wireguard preshared key: ", err)
			return peer, err
		}
	}
	peer.CreatedAt = time.Now().UTC()
	peer.UpdatedAt = peer.CreatedAt

	if err := w.store.SavePeer(peer); err != nil {
		return peer, err
	}
	err = w.applyConfig()
	if err != nil {
		return model.Peer{}, err
	}
	fmt.Println(err)
	return peer, nil
}

func (w *WireguardService) EditPeer(id string, peerValue model.Peer) (model.PeerData, error) {
	peerData, err := w.store.GetPeerByID(id, model.QRCodeSettings{Enabled: false})

	if err != nil {
		return peerData, err
	}

	server, err := w.store.GetServer()
	if err != nil {
		return peerData, err
	}

	peer := *peerData.Peer
	allocatedIPs, err := util.GetAllocatedIPs(peer.ID)

	check, err := util.ValidateIPAllocation(server.Interface.Addresses, allocatedIPs, peer.AllocatedIPs)
	if !check {
		return peerData, err
	}
	if util.ValidateAllowedIPs(peer.AllowedIPs) == false {
		logrus.Warnf("Invalid Allowed IPs input from user: %v", peer.AllowedIPs)
		return peerData, err
	}

	peer.Name = peerValue.Name
	peer.Email = peerValue.Email
	peer.Enabled = peerValue.Enabled
	peer.UseServerDNS = peerValue.UseServerDNS
	peer.AllocatedIPs = peerValue.AllocatedIPs
	peer.AllowedIPs = peerValue.AllowedIPs
	peer.UpdatedAt = time.Now().UTC()

	if err := w.store.SavePeer(peer); err != nil {
		return peerData, err
	}
	logrus.Infof("Updated client information successfully => %v", peer)

	return peerData, nil
}

func (w *WireguardService) DeletePeer(id string) error {
	fmt.Println(id)
	if err := w.store.DeletePeer(id); err != nil {
		logrus.Error("Cannot delete wireguard client: ", err)
		return err
	}
	return nil
}

func (w *WireguardService) applyConfig() error {
	server, err := w.store.GetServer()
	if err != nil {
		logrus.Error("[Server] Cannot get server config: ")
		return err
	}
	peers, err := w.store.GetPeers(false)
	if err != nil {
		logrus.Error("[Peers] Cannot get peers config")
		return err
	}
	settings, err := w.store.GetGlobalSettings()
	if err != nil {
		logrus.Error("[Peers] Cannot get peers config")
		return err
	}
	err = writeWireGuardServerConfig(server, peers, settings)
	if err != nil {
		return err
	}
	return nil
}

func writeWireGuardServerConfig(serverConfig model.Server, peersData []model.PeerData, globalSettings model.GlobalSetting) error {
	var tmplWireguardConf string
	fileContentBytes, err := os.ReadFile("./template/wg.conf")
	if err != nil {
		return err
	}
	tmplWireguardConf = string(fileContentBytes)
	// parse the template
	t, err := template.New("wg_config").Parse(tmplWireguardConf)
	if err != nil {
		return err
	}
	// write config file to disk
	f, err := os.Create(globalSettings.ConfigFilePath)
	if err != nil {
		return err
	}
	config := map[string]interface{}{
		"serverConfig":   serverConfig,
		"peersData":      peersData,
		"globalSettings": globalSettings,
	}
	err = t.Execute(f, config)
	if err != nil {
		return err
	}
	f.Close()

	return nil
}
