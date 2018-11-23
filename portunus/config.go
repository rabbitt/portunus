package portunus

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type Config struct {
	Server    ConfigServer           `mapstructure:"server" json:"server"`
	Origin    ConfigOrigin           `mapstructure:"origin" json:"origin"`
	Proxy     ConfigProxy            `mapstructure:"proxy" json:"proxy"`
	Transform ConfigHeaderTransforms `mapstructure:"transform" json:"transform"`
	Routes    map[string]ConfigRoute `mapstructure:"routes" json:"routes"`
	NewRelic  ConfigNewRelic         `mapstructure:"newrelic" json:"newrelic"`
}

func (self *Config) DumpJson() {
	json, err := json.MarshalIndent(self, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(json))
}

type ConfigOrigin struct {
	Dynamic   ConfigDynamicOrigin   `mapstructure:"dynamic" json:"dynamic"`
	Unmatched ConfigUnmatchedOrigin `mapstructure:"unmatched" json:"unmatched"`
	Timeout   time.Duration         `mapstructure:"timeout" json:"timeout"`
}

type ConfigDynamicOrigin struct {
	Host string `mapstructure:"host" json:"host"`
}

type ConfigUnmatchedOrigin struct {
	Host     string                  `mapstructure:"host" json:"host"`
	Response ConfigUnmatchedResponse `mapstructure:"response" json:"response"`
}

type ConfigUnmatchedResponse struct {
	Code int    `mapstructure:"code" json:"code"`
	Body string `mapstructure:"body" json:"body"`
}

type ConfigProxy struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	Username string `mapstructure:"username" json:"username"`
	Password string `mapstructure:"password" json:"password"`
}

type ConfigServer struct {
	BindAddress string `mapstructure:"bind_address" json:"bind_address"`
	Threads     int    `mapstructure:"threads" json:"threads"`
	TlsEnabled  bool   `mapstructure:"tls_enabled" json:"tls_enabled"`
	TlsCert     string `mapstructure:"tls_cert" json:"tls_cert"`
	TlsKey      string `mapstructure:"tls_key" json:"tls_key"`
}
type ConfigHeaderTransforms struct {
	Request  ConfigHeaderTransform `mapstructure:"request" json:"request"`
	Response ConfigHeaderTransform `mapstructure:"response" json:"response"`
}

type ConfigHeaderTransform struct {
	Insert   map[string]string `mapstructure:"insert" json:"insert"`
	Override map[string]string `mapstructure:"override" json:"override"`
	Delete   []string          `mapstructure:"delete" json:"delete"`
}

func (self *ConfigHeaderTransform) Empty() bool {
	return (self.Insert == nil || len(self.Insert) == 0) &&
		(self.Override == nil || len(self.Override) == 0) &&
		(self.Delete == nil || len(self.Delete) == 0)
}

type ConfigRoute struct {
	Paths []string `mapstructure:"paths" json:"paths"`
}

type ConfigNewRelic struct {
	Enabled    bool   `mapstructure:"enabled" json:"enabled"`
	AppName    string `mapstructure:"app_name" json:"app_name"`
	LicenseKey string `mapstructure:"license_key" json:"license_key"`
}
