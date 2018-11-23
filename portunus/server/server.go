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
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httputil"
	"runtime"
	"time"

	log "github.com/rabbitt/portunus/portunus/logging"
	config "github.com/spf13/viper"
	golog "log"
)

type Server struct {
	router    *Router
	proxy     *httputil.ReverseProxy
	transport *http.Transport
	startup   time.Time
}

func NewServer() *Server {
	server := &Server{}
	server.router = NewRouter(server)
	server.startup = time.Now()
	log.InfoWithFields("Portunus starting", log.Fields{"bind_address": config.GetString("server.bind_address")})

	return server
}

func (self *Server) Run() int {
	runtime.GOMAXPROCS(int(config.GetInt("server.threads")))

	// Setup the logWriter for the proxy service (see: server#proxyHandler)
	proxyLogWriter := log.Writer()
	defer proxyLogWriter.Close()

	self.proxy = &httputil.ReverseProxy{
		Director:  func(req *http.Request) {}, // header rewrites are handled in proxyTransport
		Transport: &proxyTransport{self},
		ErrorLog:  golog.New(proxyLogWriter, "", 0),
	}

	self.transport = &http.Transport{
		Proxy: nil, // No proxying of upstream requests
		DialContext: (&net.Dialer{
			Timeout:   config.GetDuration("network.timeouts.connect") * time.Second,
			KeepAlive: config.GetDuration("network.timeouts.keepalive") * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          config.GetInt("network.max_idle_connections"),
		MaxIdleConnsPerHost:   config.GetInt("network.max_idle_per_host"),
		IdleConnTimeout:       config.GetDuration("network.timeouts.idle_connection") * time.Second,
		TLSHandshakeTimeout:   config.GetDuration("network.timeouts.tls_handshake") * time.Second,
		ExpectContinueTimeout: config.GetDuration("network.timeouts.continue") * time.Second,
	}

	if config.GetBool("server.tls.enabled") {
		cfg := &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		}

		srv := &http.Server{
			Addr:         config.GetString("server.bind_address"),
			ReadTimeout:  config.GetDuration("network.timeouts.read") * time.Second,
			WriteTimeout: config.GetDuration("network.timeouts.write") * time.Second,
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
			TLSConfig:    cfg,
			Handler:      self.router.mux,
		}

		log.Fatal(srv.ListenAndServeTLS(
			config.GetString("server.tls.cert"), config.GetString("server.tls.key")))
	} else {

		srv := &http.Server{
			Addr:         config.GetString("server.bind_address"),
			ReadTimeout:  config.GetDuration("network.timeouts.read") * time.Second,
			WriteTimeout: config.GetDuration("network.timeouts.write") * time.Second,
			Handler:      self.router.mux,
		}

		log.Fatal(srv.ListenAndServe())
	}

	return 0
}
