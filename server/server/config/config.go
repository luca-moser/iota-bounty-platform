package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config interface{}

const configLoc = "./configs/app.json"

func LoadConfig() (*Configuration, error) {
	conf := &Configuration{}
	configBytes, err := ioutil.ReadFile(configLoc)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(configBytes, conf); err != nil {
		return nil, err
	}
	return conf, err
}

type Configuration struct {
	Name      string
	Dev       bool
	Verbose   bool
	GitHub    GitHubConfig
	Account   AccountConfig
	HTTP      WebConfig
	DB        DBConfig
}

type GitHubConfig struct {
	AuthToken string `json:"auth_token"`
	WebHook   struct {
		URL           string
		ListenAddress string `json:"listen_address"`
		URLPath       string `json:"url_path"`
		TLS           bool
	} `json:"web_hook"`
	SyncIntervalSeconds int `json:"sync_interval_seconds"`
}

type AccountConfig struct {
	Node                       string `json:"node"`
	Collection                 string `json:"collection"`
	MWM                        uint64 `json:"mwm"`
	GTTADepth                  uint64 `json:"gtta_depth"`
	SecurityLevel              uint64 `json:"security_level"`
	NTPServer                  string `json:"ntp_server"`
}

type DBConfig struct {
	URI      string `json:"uri"`
	DBName   string `json:"dbname"`
	CollName string `json:"collname"`
}

type WebConfig struct {
	Domain        string
	ListenAddress string `json:"listen_address"`
	Assets        struct {
		Static  string
		HTML    string
		Favicon string
	}
	LogRequests bool
}
