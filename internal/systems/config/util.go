package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// loadFromDisk reads and unmarshals config.json or returns defaults.
func loadFromDisk(path string) (*RootConfig, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		log.Printf("[Config] %s not found — creating default.\n", path)
		cfg := DefaultConfig()
		if err := saveToDisk(path, cfg); err != nil {
			return cfg, err
		}
		return cfg, nil
	} else if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var cfg RootConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// saveToDisk writes config as JSON (overwrites file).
func saveToDisk(path string, cfg *RootConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

