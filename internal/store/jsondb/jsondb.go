package jsondb

import (
	"github.com/sdomino/scribble"
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
