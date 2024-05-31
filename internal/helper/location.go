package helper

import (
	"fmt"
	"stocms/internal/config"
	"strconv"
)

// GetHost constructs the URL based on the given domain and config, with an optional port.
func GetHost(conf *config.Config, domain string, ports ...int) string {
	port := getPort(conf, ports...)
	return buildURL(conf.Protocol, domain, port)
}

// getPort retrieves the port number from the config or the optional ports parameter.
func getPort(conf *config.Config, ports ...int) string {
	if len(ports) > 0 {
		return strconv.Itoa(ports[0])
	} else if conf.Port != 0 {
		return strconv.Itoa(conf.Port)
	}
	return ""
}

// buildURL constructs the URL string based on the protocol, domain, and optional port.
func buildURL(protocol, domain, port string) string {
	if port != "" {
		return fmt.Sprintf("%v://%v:%v", protocol, domain, port)
	}
	return fmt.Sprintf("%v://%v", protocol, domain)
}
