package utils

import "sync"

var configMu sync.Mutex

// lockConfig serialises LoadConfig: viper's defaults map is not safe for concurrent writes, and the auth
// middleware loads the config on every request. Use as: defer lockConfig()().
func lockConfig() func() {
	configMu.Lock()
	return configMu.Unlock
}
