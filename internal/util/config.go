package util

const (
	DefaultServerAddress       = "10.252.1.0/24"
	DefaultServerPort          = 51820
	DefaultDNS                 = "1.1.1.1"
	DefaultMTU                 = 1450
	DefaultPersistentKeepalive = 15
	DefaultConfigFilePath      = "/etc/wireguard/wg0.conf"
	DefaultForwardMark         = "0xca6c"
	ServerAddressesEnvVar      = "WG_SERVER_INTERFACE_ADDRESSES"
	ServerListenPortEnvVar     = "WG_SERVER_LISTEN_PORT"
	ServerPostUpScriptEnvVar   = "WG_SERVER_POST_UP_SCRIPT"
	ServerPostDownScriptEnvVar = "WG_SERVER_POST_DOWN_SCRIPT"
	EndpointAddressEnvVar      = "WG_ENDPOINT_ADDRESS"
	DNSEnvVar                  = "WG_DNS"
	MTUEnvVar                  = "WG_MTU"
	PersistentKeepaliveEnvVar  = "WG_PERSISTENT_KEEPALIVE"
	ForwardMarkEnvVar          = "WG_FORWARD_MARK"
	ConfigFilePathEnvVar       = "WG_CONFIG_FILE_PATH"
)
