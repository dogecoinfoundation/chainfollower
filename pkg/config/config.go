package config

import "github.com/BurntSushi/toml"

type Config struct {
	Path    string
	RpcUrl  string `toml:"rpc_url"`
	RpcUser string `toml:"rpc_user"`
	RpcPass string `toml:"rpc_pass"`
	ZmqUrl  string `toml:"zmq_url"`
	DbUrl   string `toml:"db_url"`
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
