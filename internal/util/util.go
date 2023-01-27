package util

import (
	"fmt"
	externalip "github.com/glendc/go-external-ip"
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
