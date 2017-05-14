package YAST

import (
	"encoding/json"
	"os"
)

type Config struct {
	RemoteURL       string `json:"remote_url"`
	DbType          string `json:"db_type"`
	DbSrc           string `json:"db_src"`
	LocalURL        string `json:"local_url"`
	UpdaterInterval int    `json:"updater_interval"`
}

func Loadconfig(str string) *Config {
	configFile, err := os.Open(str)
	if err != nil {
		panic(err.Error())
	}
	decoder := json.NewDecoder(configFile)
	config := Config{}
	if err = decoder.Decode(&config); err != nil {
		panic(err.Error())
	}
	return &config
}
