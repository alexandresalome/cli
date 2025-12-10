package fleet

import (
	"fmt"
	"os"
)

type FleetCache struct {
	directory string
}

func NewFleetCache(directory string) *FleetCache {
	return &FleetCache{
		directory: directory,
	}
}

func (c *FleetCache) Write(key string, data []byte) error {
	path := fmt.Sprintf("%s/%s.cache", c.directory, key)
	return os.WriteFile(path, data, 0644)
}

func (c *FleetCache) Read(key string) ([]byte, error) {
	path := fmt.Sprintf("%s/%s.cache", c.directory, key)
	return os.ReadFile(path)
}
