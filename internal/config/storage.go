package config

import "ncobase/pkg/storage"

// getStorageConfig get storage config
func getStorageConfig() storage.Config {
	return storage.Config{
		Provider: c.GetString("storage.provider"),
		ID:       c.GetString("storage.id"),
		Secret:   c.GetString("storage.secret"),
		Region:   c.GetString("storage.region"),
		Bucket:   c.GetString("storage.bucket"),
		Endpoint: c.GetString("storage.endpoint"),
	}
}
