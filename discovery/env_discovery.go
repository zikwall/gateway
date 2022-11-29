package discovery

import (
	"os"
	"strings"
	"sync"
)

type Environment struct {
	mu       sync.RWMutex
	services map[string]string
}

func (d *Environment) Lookup(name string) (string, bool) {
	d.mu.RLock()
	dest, ok := d.services[name]
	d.mu.RUnlock()
	if ok {
		return dest, ok
	}
	upName := strings.ToUpper(name)
	hostEnv := upName + "_SERVICE_HOST"
	portEnv := upName + "_SERVICE_PORT"
	host := os.Getenv(hostEnv)
	port := os.Getenv(portEnv)
	if host == "" {
		return "", false
	}
	dest = host
	if port != "" {
		dest = host + ":" + port
	}
	d.mu.Lock()
	d.services[name] = dest
	d.mu.Unlock()
	return dest, true
}

func NewEnvironment() *Environment {
	d := &Environment{
		services: map[string]string{},
	}
	return d
}
