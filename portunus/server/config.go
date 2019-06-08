package server

import (
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/r3labs/diff"
	log "github.com/rabbitt/portunus/portunus/logging"
	"github.com/sanity-io/litter"
)

type ConfigDNS struct {
	Resolvers []string `mapstructure:"resolvers" diff:"resolvers"`
}

type ConfigLogging struct {
	Level string `mapstructure:"level" diff:"level"`
}

type ConfigNetworkTimeouts struct {
	Connect        time.Duration `mapstructure:"connect" diff:"connect"`
	Continue       time.Duration `mapstructure:"continue" diff:"continue"`
	IdleConnection time.Duration `mapstructure:"idle_connection" diff:"idle_connection"`
	Keepalive      time.Duration `mapstructure:"keepalive" diff:"keepalive"`
	TLSHandshake   time.Duration `mapstructure:"tls_handshake" diff:"tls_handshake"`
	Read           time.Duration `mapstructure:"read" diff:"read"`
	Write          time.Duration `mapstructure:"write" diff:"write"`
}

type ConfigNetwork struct {
	MaxIdleConnections int                   `mapstructure:"max_idle_connections" diff:"max_idle_connections"`
	MaxIdlePerHost     int                   `mapstructure:"max_idle_per_host" diff:"max_idle_per_host"`
	Timeouts           ConfigNetworkTimeouts `mapstructure:"timeouts" diff:"timeouts"`
}

type ConfigErrorCollections struct {
	Enabled           bool  `mapstructure:"enabled" diff:"enabled"`
	IgnoreStatusCodes []int `mapstructure:"ignore_status_codes" diff:"ignore_status_codes"`
}

type ConfigNewRelic struct {
	AppName         string                 `mapstructure:"app_name" diff:"app_name"`
	Enabled         bool                   `mapstructure:"enabled" diff:"enabled"`
	ErrorCollector  ConfigErrorCollections `mapstructure:"error_collector" diff:"error_collector"`
	HighSecurity    bool                   `mapstructure:"high_security" diff:"high_security"`
	HostDisplayName string                 `mapstructure:"host_display_name" diff:"host_display_name"`
	Labels          map[string]string      `mapstructure:"labels" diff:"labels"`
	LicenseKey      string                 `mapstructure:"license_key" diff:"license_key"`
	ProxyURL        string                 `mapstructure:"proxy_url" diff:"proxy_url"`
}

type ConfigResponseEntry struct {
	Body string `mapstructure:"body" diff:"body"`
	Code int    `mapstructure:"code" diff:"code"`
}

type ConfigResponse struct {
	NotFound    ConfigResponseEntry `mapstructure:"not_found" diff:"not_found"`
	ServerError ConfigResponseEntry `mapstructure:"server_error" diff:"server_error"`
}

type ConfigRoute struct {
	Upstream                 string   `mapstructure:"upstream" diff:"upstream"`
	Paths                    []string `mapstructure:"paths" diff:"paths"`
	AggregateChunkedRequests bool     `mapstructure:"aggregate_chunked_requests" diff:"aggregate_chunked_requests"`
}

type ConfigTLS struct {
	Enabled bool   `mapstructure:"enabled" diff:"enabled"`
	Cert    string `mapstructure:"cert" diff:"cert"`
	Key     string `mapstructure:"key" diff:"key"`
}

type ConfigHTTP2 struct {
	Enabled bool `mapstructure:"enabled" diff:"enabled"`
}
type ConfigServer struct {
	BindAddress     string        `mapstructure:"bind_address" diff:"bind_address"`
	Threads         int           `mapstructure:"threads" diff:"threads"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" diff:"shutdown_timeout" `
	HTTP2           ConfigHTTP2   `mapstructure:"http2" diff:"http2"`
	TLS             ConfigTLS     `mapstructure:"tls" diff:"tls"`
}

type ConfigTransformEntry struct {
	Insert   map[string]string `mapstructure:"insert" diff:"insert"`
	Override map[string]string `mapstructure:"override" diff:"override"`
	Delete   []string          `mapstructure:"delete" diff:"delete"`
}

type ConfigTransform struct {
	Request  ConfigTransformEntry `mapstructure:"request" diff:"request"`
	Response ConfigTransformEntry `mapstructure:"response" diff:"response"`
}

type Config struct {
	ConfigFile string                 `mapstructure:"config" diff:"config"`
	DNS        ConfigDNS              `mapstructure:"dns" diff:"dns"`
	Logging    ConfigLogging          `mapstructure:"logging" diff:"logging"`
	Network    ConfigNetwork          `mapstructure:"network" diff:"network"`
	NewRelic   ConfigNewRelic         `mapstructure:"newrelic" diff:"newrelic"`
	Response   ConfigResponse         `mapstructure:"response" diff:"response"`
	Routes     map[string]ConfigRoute `mapstructure:"routes" diff:"routes"`
	Server     ConfigServer           `mapstructure:"server" diff:"server"`
	Transform  ConfigTransform        `mapstructure:"transform" diff:"transform"`
}

func DecodeConfigMap(input map[string]interface{}) (*Config, error) {
	var newConfig Config

	if log.IsTraceEnabled() {
		litter.Dump(input)
	}

	decodeConfig := &mapstructure.DecoderConfig{
		DecodeHook:       mapstructure.StringToTimeDurationHookFunc(),
		WeaklyTypedInput: true,
		Result:           &newConfig,
	}

	decoder, err := mapstructure.NewDecoder(decodeConfig)
	if err != nil {
		return nil, err
	}

	err = decoder.Decode(input)
	if err != nil {
		return nil, err
	}

	return &newConfig, nil
}

func (c *Config) LoadFromMap(input map[string]interface{}, onChangeFunc func(old, new *Config)) {
	newConfig, err := DecodeConfigMap(input)
	if err != nil {
		log.Error(err)
	}

	onChangeFunc(c, newConfig)

	*c = *newConfig

	log.SetLogLevel(log.GetLogger(), c.Logging.Level)
	SetResolvers(c.DNS.Resolvers)

	if log.IsTraceEnabled() {
		litter.Dump(c)
	}
}

func (c *Config) LoadAndLogDiff(input map[string]interface{}) {
	c.LoadFromMap(input, func(old, new *Config) {
		changelog, err := diff.Diff(*old, *new)
		if err == nil {
			log.InfoWithFields("Configuration Reloaded", log.Fields{"changelog": changelog})
		}
	})
}

// Settings contains Viper loaded settings
var Settings = &Config{}
