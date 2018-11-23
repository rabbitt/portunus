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
	"bytes"
	"expvar"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/codahale/metrics"
	"github.com/fanyang01/radix"
	newrelic "github.com/newrelic/go-agent"
	log "github.com/rabbitt/portunus/portunus/logging"
	config "github.com/spf13/viper"
	"github.com/zenazn/goji/web/mutil"
)

var (
	requests  = metrics.Counter("HTTP.Requests")
	responses = metrics.Counter("HTTP.Responses")

	// a five-minute window tracking 1ms-3min
	latency = metrics.NewHistogram("HTTP.Latency", 1, 1000*60*3, 3)
)

type Router struct {
	mux       *http.ServeMux
	server    *Server
	routeTree *routeTree
}

func NewRouter(server *Server) *Router {
	router := &Router{
		mux:       http.NewServeMux(),
		server:    server,
		routeTree: NewRouteTree().Load(),
	}
	router.setupRoutes()
	return router
}

func (self *Router) setupRoutes() {
	// Metrics should be locked down by auth, or some other mechanism
	self.mux.HandleFunc("/__portunus_metrics__", logRequest(expvarHandler()))
	self.mux.HandleFunc("/__portunus_ping__", aliveHandler())

	proxyHandlerFunc := logRequest(metricHandler(self.server.proxyHandler()))
	if config.GetBool("newrelic.enabled") {
		self.mux.HandleFunc(newrelic.WrapHandleFunc(nrApp, "/", proxyHandlerFunc))
	} else {
		self.mux.HandleFunc("/", proxyHandlerFunc)
	}
}

func (self *Server) proxyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.DebugWithFields("Request Received", log.Fields{
			"remote.addr":  r.RemoteAddr,
			"request.host": r.Host,
			"request.uri":  r.RequestURI,
			"user.agent":   r.UserAgent(),
		})

		self.proxy.ServeHTTP(w, r)
	}
}

func aliveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")

		if r.Method == "GET" {
			io.WriteString(w, "pong")
		}

		return
	}
}

func logRequest(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrappedWriter := mutil.WrapWriter(w)
		h(wrappedWriter, r)
		log.InfoWithFields("Request Handled", log.Fields{
			"remote.address":     r.RemoteAddr,
			"request.host":       r.Host,
			"request.method":     r.Method,
			"request.uri":        r.RequestURI,
			"request.proto":      r.Proto,
			"response.status":    wrappedWriter.Status(),
			"response.bytes":     wrappedWriter.BytesWritten(),
			"request.user-agent": r.Header.Get("User-Agent"),
			"request.duration":   time.Since(start),
		})
	}
}

func expvarHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(w, "{\n")
		first := true
		expvar.Do(func(kv expvar.KeyValue) {
			if !first {
				fmt.Fprintf(w, ",\n")
			}
			first = false
			fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
		})
		fmt.Fprintf(w, "\n}\n")
	}
}

func metricHandler(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requests.Add()
		defer responses.Add()
		defer func(start time.Time) {
			_ = latency.RecordValue(int64(time.Since(start).Seconds() * 1000.0))
		}(time.Now())

		h(w, r)
	}
}

type Route struct {
	Name         string
	MatchedPath  string
	Upstream     string
	AggReqChunks bool
}

func (r *Route) AggregateRequestChunks() bool {
	return r.AggReqChunks
}

type routeTree struct {
	mutex sync.Mutex
	radix *radix.PatternTrie
}

func NewRouteTree() *routeTree {
	return &routeTree{}
}

func normalizePath(path string) string {
	buffer := bytes.NewBufferString(`/`)
	buffer.WriteString(strings.TrimLeft(path, "/"))
	return buffer.String()
}

func (self *routeTree) Load() *routeTree {
	start := time.Now()
	defer func() {
		log.DebugWithFields("routeTree Loaded", log.Fields{"duration": time.Since(start)})
	}()

	self.mutex.Lock()
	defer self.mutex.Unlock()
	self.radix = radix.NewPatternTrie()

	routes := config.Sub("routes")
	for name, _ := range config.GetStringMap("routes") {
		routeConfig := routes.Sub(fmt.Sprintf("%s", name))
		routeAggReq := routeConfig.GetBool("aggregate_chunked_requests")
		routeUpstream := routeConfig.GetString("upstream")

		for _, path := range routeConfig.GetStringSlice("paths") {
			route := &Route{
				Name:         name,
				MatchedPath:  path,
				Upstream:     routeUpstream,
				AggReqChunks: routeAggReq,
			}

			self.radix.Add(normalizePath(path), route)

			log.DebugWithFields("Added Route", log.Fields{
				"route.name":                       name,
				"route.path":                       path,
				"route.upstream":                   routeUpstream,
				"route.aggregate_chunked_requests": routeAggReq,
			})
		}
	}

	return self
}

func (self *routeTree) insert(s string, route *Route) (*Route, bool) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if v, has := self.radix.Add(normalizePath(s), route); has {
		oldRoute := v.(*Route)
		return oldRoute, true
	}
	return nil, false
}

func (self *routeTree) Lookup(s string) (*Route, bool) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if v, ok := self.radix.Lookup(s); ok {
		return v.(*Route), true
	}

	return nil, false
}
