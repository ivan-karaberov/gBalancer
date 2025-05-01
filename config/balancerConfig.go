package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
)

// Config представляет структуру конфигурационного файла.
// Содержит настройки порта, серверов-бэкендов
type BalancerConfig struct {
	Port     int16    `json:"port"`     // Порт, на котором будет запущен сервер.
	Backends []string `json:"backends"` // Список адресов серверов-бэкендов.
}

// Init инициализирует конфигурацию, загружая её из файла
func ReadConfig(filename string) (*BalancerConfig, error) {
	var balancerConfig BalancerConfig

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed open file with filename %s: %w", filename, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&balancerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed decode json %s", err.Error())
	}

	if balancerConfig.Backends == nil {
		return nil, errors.New("servers list is empty")
	}
	if balancerConfig.Port == 0 {
		return nil, errors.New("listening port not specified")
	}

	log.Printf("Data successfully retrieved from %s", filename)
	return &balancerConfig, nil
}
