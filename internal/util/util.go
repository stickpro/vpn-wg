package util

import (
	"encoding/json"
	"fmt"
	externalip "github.com/glendc/go-external-ip"
	"github.com/sdomino/scribble"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"vpn-wg/internal/model"
)

func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func LookupEnvOrStrings(key string, defaultVal []string) []string {
	if val, ok := os.LookupEnv(key); ok {
		return strings.Split(val, ",")
	}
	return defaultVal
}

func LookupEnvOrInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			fmt.Fprintf(os.Stderr, "LookupEnvOrInt[%s]: %v\n", key, err)
		}
		return v
	}
	return defaultVal
}

func GetPublicIP() (model.Interface, error) {
	cfg := externalip.ConsensusConfig{}
	cfg.Timeout = time.Second * 5
	consensus := externalip.NewConsensus(&cfg, nil)

	// add trusted voters
	consensus.AddVoter(externalip.NewHTTPSource("http://checkip.amazonaws.com/"), 1)
	consensus.AddVoter(externalip.NewHTTPSource("http://whatismyip.akamai.com"), 1)
	consensus.AddVoter(externalip.NewHTTPSource("http://ifconfig.top"), 1)

	publicInterface := model.Interface{}
	publicInterface.Name = "Public Address"

	ip, err := consensus.ExternalIP()
	if err != nil {
		publicInterface.IPAddress = "N/A"
	}
	publicInterface.IPAddress = ip.String()

	return publicInterface, err
}

func GetAllocatedIPs(ignorepeerID string) ([]string, error) {
	allocatedIPs := make([]string, 0)
	// initialize database directory TODO change normal init
	dir := "./db"
	db, err := scribble.New(dir, nil)
	if err != nil {
		return nil, err
	}
	// read server information
	serverInterface := model.ServerInterface{}
	if err := db.Read("server", "interfaces", &serverInterface); err != nil {
		return nil, err
	}
	// append server's addresses to the result
	for _, cidr := range serverInterface.Addresses {
		ip, err := GetIPFromCIDR(cidr)
		if err != nil {
			return nil, err
		}
		allocatedIPs = append(allocatedIPs, ip)
	}
	// read peer information
	records, err := db.ReadAll("peers")
	if err != nil {
		return nil, err
	}
	// append peer's addresses to the result
	for _, f := range records {
		peer := model.Peer{}
		if err := json.Unmarshal([]byte(f), &peer); err != nil {
			return nil, err
		}

		if peer.ID != ignorepeerID {
			for _, cidr := range peer.AllocatedIPs {
				ip, err := GetIPFromCIDR(cidr)
				if err != nil {
					return nil, err
				}
				allocatedIPs = append(allocatedIPs, ip)
			}
		}
	}

	return allocatedIPs, nil
}

func GetIPFromCIDR(cidr string) (string, error) {
	ip, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}
	return ip.String(), nil
}

func ValidateIPAllocation(serverAddresses []string, ipAllocatedList []string, ipAllocationList []string) (bool, error) {
	for _, peerCIDR := range ipAllocationList {
		ip, _, _ := net.ParseCIDR(peerCIDR)

		// peerCIDR must be in CIDR format
		if ip == nil {
			return false, fmt.Errorf("Invalid ip allocation input %s. Must be in CIDR format", peerCIDR)
		}

		// return false immediately if the ip is already in use (in ipAllocatedList)
		for _, item := range ipAllocatedList {
			if item == ip.String() {
				return false, fmt.Errorf("IP %s already allocated", ip)
			}
		}

		// even if it is not in use, we still need to check if it
		// belongs to a network of the server.
		var isValid bool = false
		for _, serverCIDR := range serverAddresses {
			_, serverNet, _ := net.ParseCIDR(serverCIDR)
			if serverNet.Contains(ip) {
				isValid = true
				break
			}
		}

		// current ip allocation is valid, check the next one
		if isValid {
			continue
		} else {
			return false, fmt.Errorf("IP %s does not belong to any network addresses of WireGuard server", ip)
		}
	}

	return true, nil
}

// ValidateCIDR to validate a network CIDR
func ValidateCIDR(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return true
}

// ValidateCIDRList to validate a list of network CIDR
func ValidateCIDRList(cidrs []string, allowEmpty bool) bool {
	for _, cidr := range cidrs {
		if allowEmpty {
			if len(cidr) > 0 {
				if ValidateCIDR(cidr) == false {
					return false
				}
			}
		} else {
			if ValidateCIDR(cidr) == false {
				return false
			}
		}
	}
	return true
}

// ValidateAllowedIPs to validate allowed ip addresses in CIDR format
func ValidateAllowedIPs(cidrs []string) bool {
	if ValidateCIDRList(cidrs, false) == false {
		return false
	}
	return true
}

// ValidateExtraAllowedIPs to validate extra Allowed ip addresses, allowing empty strings
func ValidateExtraAllowedIPs(cidrs []string) bool {
	if ValidateCIDRList(cidrs, true) == false {
		return false
	}
	return true
}

func BuildPeerConfig(peer model.Peer, server model.Server, setting model.GlobalSetting) string {
	// Interface section
	peerAddress := fmt.Sprintf("Address = %s\n", strings.Join(peer.AllocatedIPs, ","))
	peerPrivateKey := fmt.Sprintf("PrivateKey = %s\n", peer.PrivateKey)
	peerDNS := ""
	if peer.UseServerDNS {
		peerDNS = fmt.Sprintf("DNS = %s\n", strings.Join(setting.DNSServers, ","))
	}
	peerMTU := ""
	if setting.MTU > 0 {
		peerMTU = fmt.Sprintf("MTU = %d\n", setting.MTU)
	}

	// Peer section
	peerPublicKey := fmt.Sprintf("PublicKey = %s\n", server.KeyPair.PublicKey)
	peerPresharedKey := ""
	if peer.PresharedKey != "" {
		peerPresharedKey = fmt.Sprintf("PresharedKey = %s\n", peer.PresharedKey)
	}

	peerAllowedIPs := fmt.Sprintf("AllowedIPs = %s\n", strings.Join(peer.AllowedIPs, ","))

	desiredHost := setting.EndpointAddress
	desiredPort := server.Interface.ListenPort
	if strings.Contains(desiredHost, ":") {
		split := strings.Split(desiredHost, ":")
		desiredHost = split[0]
		if n, err := strconv.Atoi(split[1]); err == nil {
			desiredPort = n
		} else {
			logrus.Error("Endpoint appears to be incorrectly formatted: ", err)
		}
	}
	peerEndpoint := fmt.Sprintf("Endpoint = %s:%d\n", desiredHost, desiredPort)

	peerPersistentKeepalive := ""
	if setting.PersistentKeepalive > 0 {
		peerPersistentKeepalive = fmt.Sprintf("PersistentKeepalive = %d\n", setting.PersistentKeepalive)
	}

	forwardMark := ""
	if setting.ForwardMark != "" {
		forwardMark = fmt.Sprintf("FwMark = %s\n", setting.ForwardMark)
	}

	// build the config as string
	strConfig := "[Interface]\n" +
		peerAddress +
		peerPrivateKey +
		peerDNS +
		peerMTU +
		forwardMark +
		"\n[Peer]\n" +
		peerPublicKey +
		peerPresharedKey +
		peerAllowedIPs +
		peerEndpoint +
		peerPersistentKeepalive

	return strConfig
}
