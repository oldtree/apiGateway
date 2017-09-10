package config

var cfg *Config

type Config struct {
	Version    string
	Port       string
	EtcdConfig *EtcdConfig
}

func GetConfig() *Config {
	return cfg
}

type EtcdConfig struct {
	EtcdEndpoint      []string
	ConnectionTimeout int64
	RootDir           string
}
