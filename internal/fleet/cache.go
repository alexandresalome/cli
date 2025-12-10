package fleet

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

func basedir(path string) string {
	i := len(path) - 1
	for i >= 0 && path[i] != '/' {
		i--
	}
	if i >= 0 {
		return path[:i]
	}
	return "."
}

type FleetCache struct {
	directory string
	logger    *logrus.Entry
}

func NewFleetCache(directory string, logger *logrus.Entry) *FleetCache {
	return &FleetCache{
		directory: directory,
		logger:    logger,
	}
}

func (c *FleetCache) Write(key string, data []byte) error {
	cacheFile := fmt.Sprintf("%s/%s.cache", c.directory, key)
	cacheDir := basedir(cacheFile)
	exists := true
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		exists = false
	}
	if !exists {
		c.logger.Debugf("Creating cache directory: %s", cacheDir)
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			c.logger.Errorf("Failed to create cache directory: %v", err)

			return err
		}
	}
	c.logger.Debugf("Writing cache file: %s", cacheFile)
	return os.WriteFile(cacheFile, data, 0644)
}

func (c *FleetCache) Read(key string) []byte {
	cacheFile := fmt.Sprintf("%s/%s.cache", c.directory, key)
	c.logger.Tracef("Reading cache file: %s", cacheFile)
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return nil
	}

	result, err := os.ReadFile(cacheFile)
	if err != nil {
		c.logger.Errorf("Failed to read cache file: %v", err)
		return nil
	}

	return result
}
