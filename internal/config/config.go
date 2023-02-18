package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"time"
)

const (
	defaultHttpRWTimeout          = 10 * time.Second
	defaultHttpMaxHeaderMegabytes = 1

	EnvLocal = "local"
	Prod     = "prod"

	defaultServerAddress       = "10.252.1.0/24"
	defaultServerPort          = 51820
	defaultDNS                 = "1.1.1.1"
	defaultMTU                 = 1450
	defaultPersistentKeepalive = 15
	defaultConfigFilePath      = "/home/stickpro/vpn/wg0.conf"
	defaultForwardMark         = "0xca6c"
)

type (
	Config struct {
		HTTP   HTTPConfig
		Server ServerConfig
		Global GlobalConfig
	}

	HTTPConfig struct {
		Host               string        `env:"HTTP_HOST" env-default:"localhost"`
		Port               string        `env:"HTTP_PORT" env-default:"8080"`
		ReadTimeout        time.Duration `env:"HTTP_READ_TIMEOUT"`
		WriteTimeout       time.Duration `env:"HTTP_WRITE_TIMEOUT"`
		MaxHeaderMegabytes int           `env:"HTTP_MAX_HEADER_MEGABYTES"`
	}

	ServerConfig struct {
		Addresses string `env:"WG_SERVER_INTERFACE_ADDRESSES"`
		Port      int    `env:"WG_SERVER_LISTEN_PORT"`
		PostUp    string `env:"WG_SERVER_POST_UP_SCRIPT"`
		PostDown  string `env:"WG_SERVER_POST_DOWN_SCRIPT"`
	}

	GlobalConfig struct {
		Addresses           string `env:"WG_ENDPOINT_ADDRESS"`
		DNS                 string `env:"WG_DNS"`
		MTU                 int    `env:"WG_MTU"`
		PersistentKeepalive int    `env:"WG_PERSISTENT_KEEPALIVE"`
		ForwardMark         string `env:"WG_FORWARD_MARK"`
		ConfigFilePath      string `env:"WG_CONFIG_FILE_PATH"`
	}
)

func Init() (*Config, error) {
	cfg := Config{}
	populateDefaults(cfg)
	fmt.Println(cfg)
	err := cleanenv.ReadEnv(&cfg.HTTP)
	if err != nil {
		return nil, err
	}

	err = cleanenv.ReadEnv(&cfg.Global)
	if err != nil {
		return nil, err
	}

	err = cleanenv.ReadEnv(&cfg.Server)
	if err != nil {
		return nil, err
	}

	log.Println("Parsed Configuration")
	return &cfg, nil
}

func populateDefaults(cfg Config) {
	cfg.HTTP.ReadTimeout = defaultHttpRWTimeout
	cfg.HTTP.WriteTimeout = defaultHttpRWTimeout
	cfg.HTTP.MaxHeaderMegabytes = defaultHttpMaxHeaderMegabytes

	cfg.Server.Addresses = defaultServerAddress
	cfg.Server.Port = defaultServerPort
	cfg.Server.PostUp = ""
	cfg.Server.PostUp = ""

	cfg.Global.Addresses = ""
	cfg.Global.DNS = defaultDNS
	cfg.Global.MTU = defaultMTU
	cfg.Global.PersistentKeepalive = defaultPersistentKeepalive
	cfg.Global.ForwardMark = defaultForwardMark
	cfg.Global.ConfigFilePath = defaultConfigFilePath

}
