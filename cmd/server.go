// Copyright © 2018 Carl P. Corliss <carl@corliss.name>
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"os"
	signals "os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/rabbitt/portunus/portunus"
	log "github.com/rabbitt/portunus/portunus/logging"
	"github.com/rabbitt/portunus/portunus/server"
	"github.com/spf13/cobra"
	config "github.com/spf13/viper"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the Portunus service",
	Run: func(cmd *cobra.Command, args []string) {
		sig := make(chan os.Signal, 1)
		signals.Notify(sig, syscall.SIGUSR1, syscall.SIGHUP)

		// Initial load of Server.Config uses LoadFromMap
		server.Settings.LoadFromMap(config.AllSettings(), func(a, b *server.Config) {})

		go func() {

			for {
				signal := <-sig

				switch signal {
				case syscall.SIGHUP:
					if server.Settings.ConfigFile != "" {
						log.Infof("Reloading configuration on SIGHUP - yaaay")
						loadConfig()
						// Further Server.Config reloads log a diff of changes
						server.Settings.LoadAndLogDiff(config.AllSettings())
					} else {
						log.Warnf("Received config reload request when no config provided")
					}
				default:
					log.Warnf("unhandled signal: %+v", signal)
				}
			}
		}()

		server.NewServer().Run()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	cobra.OnInitialize(loadConfig)

	initFlags()
	initEnvVars()
}

// loadConfig reads in config file and ENV variables if set.
func loadConfig() {
	cfgFile := config.GetString("config")
	if cfgFile != "" {
		config.SetConfigFile(cfgFile)
	} else {
		config.SetConfigName(portunus.ConfigName)
		for _, path := range portunus.ConfigSearchPaths {
			config.AddConfigPath(path)
		}
	}

	// If a config file is found, read it in.
	if err := config.ReadInConfig(); err == nil {
		config.Set("config", config.ConfigFileUsed())
		log.InfoWithFields("Config file loaded", log.Fields{
			"config.file": config.ConfigFileUsed(),
		})
	}
}

func initFlags() {
	createFlags()
	bindFlags()
	setDefaults()
}

func createFlags() {
	serverCmd.Flags().StringP("config", "c", "", "config file")

	serverCmd.Flags().StringP("server.bind_address", "b", "0.0.0.0:8080", "address:port to bind to")
	serverCmd.Flags().IntP("server.threads", "t", runtime.GOMAXPROCS(0), "set the max number of system threads to use")
	serverCmd.Flags().BoolP("server.http2.enabled", "2", false, "enable http/2 requests (disabled by default)")
	serverCmd.Flags().BoolP("server.tls.enabled", "T", false, "enable TLS for server connections")
	serverCmd.Flags().DurationP("server.shutdown_timeout", "", 5*time.Second, "How long to wait for connections to finish, during shutdown, before forcibly quitting")
	serverCmd.Flags().StringP("server.tls.cert", "C", "", "path to TLS certificate")
	serverCmd.Flags().StringP("server.tls.key", "K", "", "path to TLS key")

	serverCmd.Flags().StringP("logging.level", "l", "info", "log level (e.g., error, warn, info, debug)")

	serverCmd.Flags().IntP("network.max_idle_connections", "", 5000, "maximum total number of idle connections to upstream servers")
	serverCmd.Flags().IntP("network.max_idle_per_host", "", 100, "maximum number of idle connections *per* upstream servers")
	serverCmd.Flags().DurationP("network.timeouts.connect", "", 5*time.Second, "timeout when connecting to upstream server")
	serverCmd.Flags().DurationP("network.timeouts.read", "", 10*time.Second, "timeout when reading from upstream server")
	serverCmd.Flags().DurationP("network.timeouts.write", "", 15*time.Second, "timeout when writing to upstream server")
	serverCmd.Flags().DurationP("network.timeouts.keepalive", "", 20*time.Second, "how long to keep upstream server connections for a single client")
	serverCmd.Flags().DurationP("network.timeouts.idle_connection", "", 90*time.Second, "how long to keep idle (unused) connections in the pool")
	serverCmd.Flags().DurationP("network.timeouts.tls_handshake", "", 5*time.Second, "timeout when during tls handshake")
	serverCmd.Flags().DurationP("network.timeouts.continue", "", 5*time.Second, "timeout while waiting for a CONTINUE request/response")

	serverCmd.Flags().StringSliceP("dns.resolvers", "r", nil, "comma separated list of dns resolvers to use (default: system resolver)")
}

func bindFlags() {
	config.BindPFlag("config", serverCmd.Flags().Lookup("config"))

	config.BindPFlag("server.bind_address", serverCmd.Flags().Lookup("server.bind_address"))
	config.BindPFlag("server.threads", serverCmd.Flags().Lookup("server.threads"))
	config.BindPFlag("server.http2.enabled", serverCmd.Flags().Lookup("server.http2.enabled"))
	config.BindPFlag("server.tls.enabled", serverCmd.Flags().Lookup("server.tls.enabled"))
	config.BindPFlag("server.shutdown_timeout", serverCmd.Flags().Lookup("server.shutdown_timeout"))
	config.BindPFlag("server.tls.cert", serverCmd.Flags().Lookup("server.tls.cert"))
	config.BindPFlag("server.tls.key", serverCmd.Flags().Lookup("server.tls.key"))

	config.BindPFlag("logging.level", serverCmd.Flags().Lookup("logging.level"))

	config.BindPFlag("network.max_idle_connections", serverCmd.Flags().Lookup("network.max_idle_connections"))
	config.BindPFlag("network.max_idle_per_host", serverCmd.Flags().Lookup("network.max_idle_per_host"))
	config.BindPFlag("network.timeouts.connect", serverCmd.Flags().Lookup("network.timeouts.connect"))
	config.BindPFlag("network.timeouts.read", serverCmd.Flags().Lookup("network.timeouts.read"))
	config.BindPFlag("network.timeouts.write", serverCmd.Flags().Lookup("network.timeouts.write"))
	config.BindPFlag("network.timeouts.keepalive", serverCmd.Flags().Lookup("network.timeouts.keepalive"))
	config.BindPFlag("network.timeouts.idle_connection", serverCmd.Flags().Lookup("network.timeouts.idle_connection"))
	config.BindPFlag("network.timeouts.tls_handshake", serverCmd.Flags().Lookup("network.timeouts.tls_handshake"))
	config.BindPFlag("network.timeouts.continue", serverCmd.Flags().Lookup("network.timeouts.continue"))

	config.BindPFlag("dns.resolvers", serverCmd.Flags().Lookup("dns.resolvers"))
}

func setDefaults() {
	config.SetDefault("server.bind_address", "0.0.0.0:8080")
	config.SetDefault("server.threads", runtime.GOMAXPROCS(0))
	config.SetDefault("server.http2.enabled", false)
	config.SetDefault("server.tls.enabled", false)
	config.SetDefault("server.shutdown_timeout", 5*time.Second)
	config.SetDefault("server.tls.cert", "")
	config.SetDefault("server.tls.key", "")

	config.SetDefault("logging.level", "info")

	config.SetDefault("network.max_idle_connections", 5000)
	config.SetDefault("network.max_idle_per_host", 100)

	config.SetDefault("network.timeouts.connect", 5*time.Second)
	config.SetDefault("network.timeouts.read", 10*time.Second)
	config.SetDefault("network.timeouts.write", 15*time.Second)
	config.SetDefault("network.timeouts.keepalive", 20*time.Second)
	config.SetDefault("network.timeouts.idle_connection", 90*time.Second)
	config.SetDefault("network.timeouts.tls_handshake", 5*time.Second)
	config.SetDefault("network.timeouts.continue", 5*time.Second)

	config.SetDefault("dns.resolvers", nil)

	config.SetDefault("routes", map[string]map[string]interface{}{
		"default": map[string]interface{}{
			"upstream":                   "{{req.host}}",
			"aggregate_chunked_requests": false,
			"paths":                      []string{"*"},
		},
	})

	config.SetDefault("transform", map[string]map[string]interface{}{
		"request": map[string]interface{}{
			"insert":   make(map[string]string, 0),
			"override": make(map[string]string, 0),
			"delete":   make([]string, 0),
		},
		"response": map[string]interface{}{
			"insert":   make(map[string]string, 0),
			"override": make(map[string]string, 0),
			"delete":   make([]string, 0),
		},
	})

	config.SetDefault("response.not_found.code", 404)
	config.SetDefault("response.not_found.body", `
    <html>
      <head>
        <title>404 - Not Found</title>
      </head>
      <body><h1>404 - Not Found</h1></body>
    </html>
  `)

	config.SetDefault("response.server_error.code", 500)
	config.SetDefault("response.server_error.body", `
    <html>
      <head>
        <title>500 - Internal Server Error</title>
      </head>
      <body>
        <h1>Internal Server Error</h1>
        <p>Please try again later.<p>
      </body>
    </html>
  `)

	config.SetDefault("newrelic.enabled", false)
	config.SetDefault("newrelic.app_name", "")
	config.SetDefault("newrelic.license_key", "")

	if hostname, err := os.Hostname(); err == nil {
		config.SetDefault("newrelic.host_display_name", hostname)
	}

	config.SetDefault("newrelic.labels", make(map[string]string, 0))
	config.SetDefault("newrelic.high_security", false)
	config.SetDefault("newrelic.error_collector.enabled", true)
	config.SetDefault("newrelic.error_collector.ignore_status_codes", []int{404})
	config.SetDefault("newrelic.proxy_url", "")
}

func initEnvVars() {
	config.SetEnvPrefix("portunus") // will be uppercased automatically
	config.AutomaticEnv()           // read in environment variables that match
}
