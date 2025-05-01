package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

// Config represents the structure of the configuration file.
type Config struct {
	Port       int16    `json:"port"`         // The port on which the server will run.
	Backends   []string `json:"backends"`     // List of backend server addresses.
	Capacity   int      `json:"capacity"`     // Capacity of the request queue.
	RatePerSec int      `json:"rate_per_sec"` // Number of requests per second.
}

// Initializes the configuration by loading it from a file
func ReadConfig(filename string) (*Config, error) {
	var config Config

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed open file with filename %s: %w", filename, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("failed decode JSOn %s", err.Error())
	}

	if config.Backends == nil {
		return nil, errors.New("servers list is empty")
	}
	if config.Port == 0 {
		return nil, errors.New("listening port not specified")
	}

	logrus.Infof("Data successfully retrieved from %s", filename)
	return &config, nil
}
