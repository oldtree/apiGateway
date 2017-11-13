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
	Version    string          `json:"version,omitempty"`
	Port       string          `json:"port,omitempty"`
	EtcdConfig *DiscoverConfig `json:"discoverconfig,omitempty"`
}

type DiscoverConfig struct {
	Name         string
	RealDiscover interface{}
}

func GetConfig() *Config {
	return cfg
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

type EtcdConfig struct {
	EtcdEndpoint      []string `json:"etcd_endpoint,omitempty"`
	ConnectionTimeout int64    `json:"connection_timeout,omitempty"`
	RootDir           string   `json:"root_dir,omitempty"`
}

type KubernetesConfig struct {
}

type BoltConfig struct {
}

type ConsulConfig struct {
}

type PostgresqlConfig struct {
}

type RedisConfig struct {
}

type LogConfig struct {
}
