package util

import (
	"fmt"
	"stocms/internal/config"
)

// Find returns the smallest index i at which x == a[i],
// or len(a) if there is no such index.
func Find(a []string, x string) int {
	for i, n := range a {
		if x == n {
			return i
		}
	}
	return len(a)
}

func FindID(a []string, x string) int {
	for i, n := range a {
		if x == n {
			return i
		}
	}
	return len(a)
}

// Contains tells whether a contains x.
func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// GetDomain returns the run domain based on the given domain and conf.
func GetDomain(conf *config.Config, domain string) string {
	if domain == "localhost" {
		return fmt.Sprintf("%v://%v:%d", conf.Protocol, conf.Domain, conf.Port)
	}
	return fmt.Sprintf("%v://%v", conf.Protocol, conf.Domain)
}
