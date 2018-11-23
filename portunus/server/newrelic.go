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

package server

import (
	"net"
	"net/http"
	"net/url"
	"time"

	newrelic "github.com/newrelic/go-agent"
	log "github.com/rabbitt/portunus/portunus/logging"
	config "github.com/spf13/viper"
)

var nrApp newrelic.Application

func init() {
	if config.GetBool("newrelic.enabled") {
		nrConfig := newrelic.NewConfig(
			config.GetString("newrelic.app_name"),
			config.GetString("newrelic.license_key"),
		)

		nrConfig.HostDisplayName = config.GetString("newrelic.host_display_name")
		nrConfig.Labels = config.GetStringMapString("newrelic.labels")
		nrConfig.HighSecurity = config.GetBool("newrelic.high_security")

		nrConfig.ErrorCollector.Enabled = config.GetBool("newrelic.error_collector.enabled")
		nrConfig.ErrorCollector.IgnoreStatusCodes = config.Get("newrelic.error_collector.ignore_status_codes").([]int)

		if nrProxyUrl := config.GetString("newrelic.proxy_url"); nrProxyUrl != "" {
			if proxy_url, err := url.Parse(nrProxyUrl); err != nil {
				log.Error(err)
				// if a proxy is configured, but not a valid URL, then we just log the error
				// and attempt a setup without the proxy
			} else {
				nrConfig.Transport = &http.Transport{
					Proxy: http.ProxyURL(proxy_url),
					DialContext: (&net.Dialer{
						Timeout:   config.GetDuration("network.timeouts.connect") * time.Second,
						KeepAlive: config.GetDuration("network.timeouts.keepalive") * time.Second,
						DualStack: true,
					}).DialContext,
					MaxIdleConns:          50, // We don't use the values from config here because we should
					MaxIdleConnsPerHost:   10, // only have a few hosts to connect to for NewRelic
					IdleConnTimeout:       config.GetDuration("network.timeouts.idle_connection") * time.Second,
					TLSHandshakeTimeout:   config.GetDuration("network.timeouts.tls_handshake") * time.Second,
					ExpectContinueTimeout: config.GetDuration("network.timeouts.continue") * time.Second,
				}
			}
		}

		var err error
		if nrApp, err = newrelic.NewApplication(nrConfig); err != nil {
			log.Error(err)
		}
	}
}
