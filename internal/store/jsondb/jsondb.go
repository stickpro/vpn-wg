package jsondb

import (
	"github.com/sdomino/scribble"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"os"
	"path"
	"time"
	"vpn-wg/internal/model"
	"vpn-wg/internal/util"
)

type JsonDB struct {
	conn   *scribble.Driver
	dbPath string
}

func New(dbPath string) (*JsonDB, error) {
	conn, err := scribble.New(dbPath, nil)
	if err != nil {
		return nil, err
	}
	ans := JsonDB{
		conn:   conn,
		dbPath: dbPath,
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
		serverInterface.Addresses = util.LookupEnvOrStrings(util.ServerAddressesEnvVar, []string{util.DefaultServerAddress})
		serverInterface.ListenPort = util.LookupEnvOrInt(util.ServerListenPortEnvVar, util.DefaultServerPort)
		serverInterface.PostUp = util.LookupEnvOrString(util.ServerPostUpScriptEnvVar, "")
		serverInterface.PostDown = util.LookupEnvOrString(util.ServerPostDownScriptEnvVar, "")
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
		endpointAddress := util.LookupEnvOrString(util.EndpointAddressEnvVar, "")

		if endpointAddress == "" {
			publicInterface, err := util.GetPublicIP()
			if err != nil {
				return err
			}
			endpointAddress = publicInterface.IPAddress
		}

		globalSetting := new(model.GlobalSetting)
		globalSetting.EndpointAddress = endpointAddress
		globalSetting.DNSServers = util.LookupEnvOrStrings(util.DNSEnvVar, []string{util.DefaultDNS})
		globalSetting.MTU = util.LookupEnvOrInt(util.MTUEnvVar, util.DefaultMTU)
		globalSetting.PersistentKeepalive = util.LookupEnvOrInt(util.PersistentKeepaliveEnvVar, util.DefaultPersistentKeepalive)
		globalSetting.ForwardMark = util.LookupEnvOrString(util.ForwardMarkEnvVar, util.DefaultForwardMark)
		globalSetting.ConfigFilePath = util.LookupEnvOrString(util.ConfigFilePathEnvVar, util.DefaultConfigFilePath)
		globalSetting.UpdatedAt = time.Now().UTC()
		o.conn.Write("server", "global_settings", globalSetting)
	}

	return nil
}
