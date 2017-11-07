package config

import (
	"encoding/json"
	"io/ioutil"
)

func init() {
	cfg = new(Config)
}

var cfg *Config

type Config struct {
	Version    string      `json:"version,omitempty"`
	Port       string      `json:"port,omitempty"`
	EtcdConfig *EtcdConfig `json:"etcd_config,omitempty"`
}

func GetConfig() *Config {
	return cfg
}

type EtcdConfig struct {
	EtcdEndpoint      []string `json:"etcd_endpoint,omitempty"`
	ConnectionTimeout int64    `json:"connection_timeout,omitempty"`
	RootDir           string   `json:"root_dir,omitempty"`
}

func LoadConfig(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}
	err = json.Unmarshal(data, cfg)
	if err != nil {
		return err
	}
	return nil
}
