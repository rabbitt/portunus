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
)

var nrApp newrelic.Application

func init() {
	if Settings.NewRelic.Enabled {
		nrConfig := newrelic.NewConfig(
			Settings.NewRelic.AppName,
			Settings.NewRelic.LicenseKey,
		)

		nrConfig.HostDisplayName = Settings.NewRelic.HostDisplayName
		nrConfig.Labels = Settings.NewRelic.Labels
		nrConfig.HighSecurity = Settings.NewRelic.HighSecurity

		nrConfig.ErrorCollector.Enabled = Settings.NewRelic.ErrorCollector.Enabled
		nrConfig.ErrorCollector.IgnoreStatusCodes = Settings.NewRelic.ErrorCollector.IgnoreStatusCodes

		if Settings.NewRelic.ProxyURL != "" {
			if proxyURL, err := url.Parse(Settings.NewRelic.ProxyURL); err != nil {
				log.Error(err)
				// if a proxy is configured, but not a valid URL, then we just log the error
				// and attempt a setup without the proxy
			} else {
				nrConfig.Transport = &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
					DialContext: (&net.Dialer{
						Timeout:   Settings.Network.Timeouts.Connect * time.Second,
						KeepAlive: Settings.Network.Timeouts.Keepalive * time.Second,
						DualStack: true,
					}).DialContext,
					MaxIdleConns:          50, // We don't use the values from config here because we should
					MaxIdleConnsPerHost:   10, // only have a few hosts to connect to for NewRelic
					IdleConnTimeout:       Settings.Network.Timeouts.IdleConnection * time.Second,
					TLSHandshakeTimeout:   Settings.Network.Timeouts.TLSHandshake * time.Second,
					ExpectContinueTimeout: Settings.Network.Timeouts.Continue * time.Second,
				}
			}
		}

		var err error
		if nrApp, err = newrelic.NewApplication(nrConfig); err != nil {
			log.Error(err)
		}
	}
}
