package helper

import (
	"fmt"
	"stocms/internal/config"
	"strconv"
)

// GetHost constructs the URL based on the given domain and config, with an optional port.
func GetHost(conf *config.Config, domain string, ports ...int) string {
	var portStr string
	if len(ports) > 0 {
		portStr = strconv.Itoa(ports[0])
	} else if conf.Port != 0 {
		portStr = strconv.Itoa(conf.Port)
	}

	if domain == "localhost" {
		return buildURL(conf.Protocol, domain, portStr)
	}
	return buildURL(conf.Protocol, domain, portStr)
}

// buildURL constructs the URL string based on the protocol, domain, and optional port.
func buildURL(protocol, domain, port string) string {
	if port == "" {
		return fmt.Sprintf("%v://%v", protocol, domain)
	}
	return fmt.Sprintf("%v://%v:%v", protocol, domain, port)
}
