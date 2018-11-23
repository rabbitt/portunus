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
	"fmt"
	"os"
	"runtime"

	"github.com/rabbitt/portunus/portunus"
	log "github.com/rabbitt/portunus/portunus/logging"
	"github.com/rabbitt/portunus/portunus/server"
	"github.com/sanity-io/litter"
	"github.com/spf13/cobra"
	config "github.com/spf13/viper"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the Portunus service",
	Run: func(cmd *cobra.Command, args []string) {
		_ = fmt.Sprintf("")
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

	if log.IsDebugEnabled() {
		litter.Dump(config.AllSettings())
		config.Debug()
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
	serverCmd.Flags().BoolP("server.tls.enabled", "T", false, "enable TLS for server connections")
	serverCmd.Flags().StringP("server.tls.cert", "C", "", "path to TLS certificate")
	serverCmd.Flags().StringP("server.tls.key", "K", "", "path to TLS key")

	serverCmd.Flags().StringP("logging.level", "l", "info", "log level (e.g., error, warn, info, debug)")

	serverCmd.Flags().IntP("network.max_idle_connections", "", 5000, "maximum total number of idle connections to upstream servers")
	serverCmd.Flags().IntP("network.max_idle_per_host", "", 100, "maximum number of idle connections *per* upstream servers")
	serverCmd.Flags().IntP("network.timeouts.connect", "", 5, "timeout when connecting to upstream server")
	serverCmd.Flags().IntP("network.timeouts.read", "", 10, "timeout when reading from upstream server")
	serverCmd.Flags().IntP("network.timeouts.write", "", 15, "timeout when writing to upstream server")
	serverCmd.Flags().IntP("network.timeouts.keepalive", "", 20, "how long to keep upstream server connections for a single client")
	serverCmd.Flags().IntP("network.timeouts.idle_connection", "", 1, "how long to keep idle (unused) connections in the pool")
	serverCmd.Flags().IntP("network.timeouts.tls_handshake", "", 5, "timeout when during tls handshake")
	serverCmd.Flags().IntP("network.timeouts.continue", "", 5, "timeout while waiting for a CONTINUE request/response")
}

func bindFlags() {
	config.BindPFlag("config", serverCmd.Flags().Lookup("config"))

	config.BindPFlag("server.bind_address", serverCmd.Flags().Lookup("server.bind_address"))
	config.BindPFlag("server.threads", serverCmd.Flags().Lookup("server.threads"))
	config.BindPFlag("server.tls.enabled", serverCmd.Flags().Lookup("server.tls.enabled"))
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
}

func setDefaults() {
	config.SetDefault("server.bind_address", "0.0.0.0:8080")
	config.SetDefault("server.threads", runtime.GOMAXPROCS(0))
	config.SetDefault("server.tls.enabled", false)
	config.SetDefault("server.tls.cert", "")
	config.SetDefault("server.tls.key", "")

	config.SetDefault("logging.level", "info")

	config.SetDefault("network.max_idle_connections", 5000)
	config.SetDefault("network.max_idle_per_host", 100)

	config.SetDefault("network.timeouts.connect", 5)
	config.SetDefault("network.timeouts.read", 10)
	config.SetDefault("network.timeouts.write", 15)
	config.SetDefault("network.timeouts.keepalive", 20)
	config.SetDefault("network.timeouts.idle_connection", 90)
	config.SetDefault("network.timeouts.tls_handshake", 5)
	config.SetDefault("network.timeouts.continue", 5)

	config.SetDefault("routes", map[string]map[string]interface{}{
		"default": map[string]interface{}{
			"upstream":                   "{{req.host}}",
			"aggregate_chunked_requests": false,
			"paths": []string{"*"},
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
