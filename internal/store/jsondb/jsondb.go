package jsondb

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/sdomino/scribble"
	"github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"os"
	"path"
	"time"
	"vpn-wg/internal/config"
	"vpn-wg/internal/model"
	"vpn-wg/internal/util"
)

type JsonDB struct {
	conn         *scribble.Driver
	dbPath       string
	configServer config.ServerConfig
	configGlobal config.GlobalConfig
}

func New(dbPath string, cfgServer config.ServerConfig, cfgGlobal config.GlobalConfig) (*JsonDB, error) {
	conn, err := scribble.New(dbPath, nil)
	if err != nil {
		return nil, err
	}
	ans := JsonDB{
		conn:         conn,
		dbPath:       dbPath,
		configServer: cfgServer,
		configGlobal: cfgGlobal,
	}
	return &ans, nil
}

func (o *JsonDB) Init() error {
	var clientPath string = path.Join(o.dbPath, "clients")
	var serverPath string = path.Join(o.dbPath, "server")

	var serverInterfacePath string = path.Join(serverPath, "interfaces.json")
	var serverKeyPairPath string = path.Join(serverPath, "keypair.json")
	var globalSettingPath string = path.Join(serverPath, "global_settings.json")

	if _, err := os.Stat(clientPath); os.IsNotExist(err) {
		os.MkdirAll(clientPath, os.ModePerm)
	}

	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		os.MkdirAll(serverPath, os.ModePerm)
	}
	// server's interface
	if _, err := os.Stat(serverInterfacePath); os.IsNotExist(err) {
		serverInterface := new(model.ServerInterface)
		serverInterface.Addresses = []string{o.configServer.Addresses}
		serverInterface.ListenPort = o.configServer.Port
		serverInterface.PostUp = o.configServer.PostUp
		serverInterface.PostDown = o.configServer.PostDown
		serverInterface.UpdatedAt = time.Now().UTC()
		o.conn.Write("server", "interfaces", serverInterface)
	}

	// server's key pair
	if _, err := os.Stat(serverKeyPairPath); os.IsNotExist(err) {
		key, err := wgtypes.GeneratePrivateKey()
		if err != nil {
			return scribble.ErrMissingCollection
		}
		serverKeyPair := new(model.ServerKeypair)
		serverKeyPair.PrivateKey = key.String()
		serverKeyPair.PublicKey = key.PublicKey().String()
		serverKeyPair.UpdatedAt = time.Now().UTC()
		o.conn.Write("server", "keypair", serverKeyPair)
	}

	if _, err := os.Stat(globalSettingPath); os.IsNotExist(err) {
		endpointAddress := o.configGlobal.Addresses

		if endpointAddress == "" {
			publicInterface, err := util.GetPublicIP()
			if err != nil {
				return err
			}
			endpointAddress = publicInterface.IPAddress
		}

		globalSetting := new(model.GlobalSetting)
		globalSetting.EndpointAddress = endpointAddress
		globalSetting.DNSServers = []string{o.configGlobal.DNS}
		globalSetting.MTU = o.configGlobal.MTU
		globalSetting.PersistentKeepalive = o.configGlobal.PersistentKeepalive
		globalSetting.ForwardMark = o.configGlobal.ForwardMark
		globalSetting.ConfigFilePath = o.configGlobal.ConfigFilePath
		globalSetting.UpdatedAt = time.Now().UTC()
		o.conn.Write("server", "global_settings", globalSetting)
	}

	return nil
}

func (o *JsonDB) GetServer() (model.Server, error) {
	server := model.Server{}
	serverInterface := model.ServerInterface{}

	if err := o.conn.Read("server", "interfaces", &serverInterface); err != nil {
		return server, err
	}
	serverKeyPair := model.ServerKeypair{}
	if err := o.conn.Read("server", "keypair", &serverKeyPair); err != nil {
		return server, err
	}
	server.Interface = &serverInterface
	server.KeyPair = &serverKeyPair
	return server, nil
}

func (o *JsonDB) GetGlobalSettings() (model.GlobalSetting, error) {
	settings := model.GlobalSetting{}
	return settings, o.conn.Read("server", "global_settings", &settings)
}

func (o *JsonDB) GetPeers(hasQRCode bool) ([]model.PeerData, error) {
	peers := []model.PeerData{}

	records, err := o.conn.ReadAll("clients")
	if err != nil {
		return peers, err
	}

	for _, f := range records {
		peer := model.Peer{}
		peersData := model.PeerData{}

		if err := json.Unmarshal([]byte(f), &peer); err != nil {
			return peers, fmt.Errorf("cannot decode client json structure: %v", err)
		}

		// generate peer qrcode image in base64
		if hasQRCode && peer.PrivateKey != "" {
			server, _ := o.GetServer()
			globalSettings, _ := o.GetGlobalSettings()

			png, err := qrcode.Encode(util.BuildPeerConfig(peer, server, globalSettings), qrcode.Medium, 256)
			if err == nil {
				peersData.QRCode = "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte(png))
			} else {
				fmt.Print("Cannot generate QR code: ", err)
			}
		}

		peersData.Peer = &peer
		peers = append(peers, peersData)
	}

	return peers, nil
}

func (o *JsonDB) SavePeer(peer model.Peer) error {
	return o.conn.Write("clients", peer.ID, peer)
}

func (o *JsonDB) GetPeerByID(peerID string, qrCodeSettings model.QRCodeSettings) (model.PeerData, error) {
	peer := model.Peer{}
	peerData := model.PeerData{}

	if err := o.conn.Read("clients", peerID, &peer); err != nil {
		logrus.Error("[Peer not found]")
		return peerData, err
	}

	if qrCodeSettings.Enabled && peer.PrivateKey != "" {
		server, _ := o.GetServer()
		globalSettings, _ := o.GetGlobalSettings()

		if !qrCodeSettings.IncludeDNS {
			globalSettings.DNSServers = []string{}
		}
		if !qrCodeSettings.IncludeMTU {
			globalSettings.MTU = 0
		}
		if !qrCodeSettings.IncludeFwMark {
			globalSettings.ForwardMark = ""
		}
		png, err := qrcode.Encode(util.BuildPeerConfig(peer, server, globalSettings), qrcode.Medium, 256)
		if err == nil {
			peerData.QRCode = "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte(png))
		} else {
			fmt.Print("Cannot generate QR code: ", err)
		}
	}
	peerData.Peer = &peer

	return peerData, nil
}

func (o *JsonDB) DeletePeer(peerID string) error {
	fmt.Println(peerID)
	return o.conn.Delete("clients", peerID)
}
