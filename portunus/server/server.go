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
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	signals "os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	golog "log"

	"github.com/coreos/go-systemd/activation"
	log "github.com/rabbitt/portunus/portunus/logging"
)

const (
	// DefaultBindingAddress is all addresses
	DefaultBindingAddress = "0.0.0.0"
	// DefaultBindingPort is standard unprivileged http
	DefaultBindingPort = 8080
)

type Server struct {
	router    *Router
	proxy     *httputil.ReverseProxy
	transport *http.Transport
	server    *http.Server
	startup   time.Time
	address   string
	finished  chan struct{}
	logger    *io.PipeWriter
}

func NewServer() *Server {
	s := &Server{}
	s.router = NewRouter(s)
	s.startup = time.Now()
	s.address = BindAddress()
	s.finished = make(chan struct{})
	s.logger = log.Writer()

	s.proxy = &httputil.ReverseProxy{
		Director:  func(req *http.Request) {}, // header rewrites are handled in proxyTransport
		Transport: NewProxyTransport(s),
		ErrorLog:  golog.New(s.logger, "", 0),
	}

	s.transport = &http.Transport{
		Proxy: nil, // No proxying of upstream requests
		DialContext: (&net.Dialer{
			Timeout:   Settings.Network.Timeouts.Connect,
			KeepAlive: Settings.Network.Timeouts.Keepalive,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          Settings.Network.MaxIdleConnections,
		MaxIdleConnsPerHost:   Settings.Network.MaxIdlePerHost,
		IdleConnTimeout:       Settings.Network.Timeouts.IdleConnection,
		TLSHandshakeTimeout:   Settings.Network.Timeouts.TLSHandshake,
		ExpectContinueTimeout: Settings.Network.Timeouts.Continue,
	}

	var tlsConfig *tls.Config
	var tlsNextProto map[string]func(*http.Server, *tls.Conn, http.Handler)

	if Settings.Server.TLS.Enabled {
		tlsConfig = &tls.Config{
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

		if !Settings.Server.HTTP2.Enabled {
			tlsNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)
		}
	}

	s.server = &http.Server{
		Addr:         Settings.Server.BindAddress,
		ReadTimeout:  Settings.Network.Timeouts.Read,
		WriteTimeout: Settings.Network.Timeouts.Write,
		Handler:      s.router.mux,
		TLSNextProto: tlsNextProto,
		TLSConfig:    tlsConfig,
	}

	return s
}

func BindAddress() (bindAddress string) {
	bindAddress = Settings.Server.BindAddress
	if bindAddress == "" {
		bindAddress = fmt.Sprintf("%s:%d", DefaultBindingAddress, DefaultBindingPort)
	} else if strings.HasPrefix(bindAddress, ":") { // e.g., ":<port>"
		bindAddress = fmt.Sprintf("%s%s", DefaultBindingAddress, bindAddress)
	} else if !strings.Contains(bindAddress, ":") { // e.g., "<ip>"
		bindAddress = fmt.Sprintf("%s:%d", bindAddress, DefaultBindingPort)
	}
	return
}

func (s *Server) Listener() (*net.Listener, error) {
	var listener net.Listener

	listeners, err := activation.Listeners()
	if err != nil {
		return nil, err
	}

	if len(listeners) > 1 {
		panic("Unexpected number of socket activation fds")
	} else if len(listeners) == 1 {
		listener = listeners[0]
	} else if len(listeners) == 0 {
		listener, err = net.Listen("tcp", s.address)
		if err != nil {
			return nil, err
		}
	}

	return &listener, nil
}

func (s *Server) ListenAndServe() {
	listener, err := s.Listener()
	if err != nil {
		// Hard fail if we can't open a socket for listening
		panic(err)
	}

	if Settings.Server.TLS.Enabled {
		log.InfoWithFields("Portunus running, listening for TLS connections", log.Fields{"bind_address": s.address})
		log.Error(s.server.ServeTLS(*listener, Settings.Server.TLS.Cert, Settings.Server.TLS.Key))
	} else {
		log.InfoWithFields("Portunus running, listening for non-TLS connections", log.Fields{"bind_address": s.address})
		log.Error(s.server.Serve(*listener))
	}
}

func (s *Server) HandleSignalShutdown() {
	log.Info("Server is shutting down")
	ctx, cancel := context.WithTimeout(context.Background(),
		Settings.Server.ShutdownTimeout*time.Second)
	defer cancel()

	s.server.SetKeepAlivesEnabled(false)
	if err := s.server.Shutdown(ctx); err != nil {
		log.Panicf("cannot gracefully shut down the server: %s", err)
	}
	close(s.finished)
	return
}

func (s *Server) HandleSignalReload() {
	return
}

func (s *Server) SetupSignalHandlers() error {

	sig := make(chan os.Signal, 1)
	signals.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP,
		syscall.SIGUSR1, syscall.SIGUSR2)

	go func() {
		for {
			signal := <-sig

			switch signal {
			case syscall.SIGTERM, syscall.SIGINT:
				s.HandleSignalShutdown()
			case syscall.SIGHUP:
				s.HandleSignalReload()
			case syscall.SIGUSR1, syscall.SIGUSR2:
				// reserved
			default:
				log.Warnf("unhandled signal: %+v", signal)
			}
		}
	}()

	return nil
}

func (s *Server) Run() int {
	defer s.logger.Close()
	runtime.GOMAXPROCS(Settings.Server.Threads)

	s.SetupSignalHandlers()
	s.ListenAndServe()

	<-s.finished

	return 0
}
