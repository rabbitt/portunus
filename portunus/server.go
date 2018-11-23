package portunus

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"runtime"
	"time"
)

type Server struct {
	config    *Config
	routeTree *RouteTree
	startup   time.Time
}

func NewServer(config *Config) *Server {
	server := &Server{
		config:    config,
		routeTree: NewRouteTree().Load(config.Routes),
	}
	return server
}

func (self *Server) transformHeaders(route *Route, httpObj interface{}) {
	var transforms ConfigHeaderTransform
	var headers *http.Header

	if req, ok := httpObj.(*http.Request); ok {
		log.Printf("Rewriting Request headers for route %+v\n", route)
		transforms = self.config.Transform.Request
		headers = &req.Header
	} else {
		resp := httpObj.(*http.Response)
		transforms = self.config.Transform.Response
		headers = &resp.Header
		log.Printf("Rewriting Response headers for route %+v\n", route)
	}

	for header, value := range transforms.Insert {
		headers.Add(header, interpolate(value, route, httpObj))
		log.Printf("Added Header [%s]: %+v\n", header, headers.Get(header))
	}

	for header, value := range transforms.Override {
		headers.Set(header, interpolate(value, route, httpObj))
		log.Printf("Overwriting Header [%s]: %+v\n", header, headers.Get(header))
	}

	for _, header := range transforms.Delete {
		log.Printf("Deleting Header [%s]:\n", header)
		headers.Del(header)
	}
}

func (self *Server) handlePingCheck(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain")
	writer.Header().Set("X-Uptime", fmt.Sprintf("%s", time.Since(self.startup)))

	if request.Method == "GET" {
		fmt.Fprint(writer, "pong")
	}

	return
}

func normalizeScheme(scheme string) string {
	if scheme == "https" {
		return "https"
	}
	return "http"
}

func copyHeaders(src, dst http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func (self *Server) getOrigin(route *Route, req *http.Request) string {
	// XXX: perform a DNS lookup of the dynamic origin to ensure it's resolvable.
	// If it's not, return the Unmatched Host and log an error.
	// Note: DNS lookups might require a caching mechanism
	if route != nil {
		return interpolate(self.config.Origin.Dynamic.Host, route, req)
	} else {
		return interpolate(self.config.Origin.Unmatched.Host, route, req)
	}
}

type proxyTransport struct {
	*Server
}

func (self *proxyTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	transport := &http.Transport{
		Proxy: nil, // FIXME: allow disabling, setting, or defaulting to environment, through configuration
		Dial: (&net.Dialer{
			Timeout: self.config.Origin.Timeout * time.Second,
		}).Dial,
	}

	// Determine the origin to proxy the request to
	start := time.Now()
	route, _ := self.routeTree.Lookup(request.URL.Path)
	origin, err := url.Parse(self.getOrigin(route, request))
	if err != nil {
		log.Fatal(err)
	}

	// keep a copy so it's availble for response rewriting
	var reqHeaders = make(http.Header)
	copyHeaders(request.Header, reqHeaders)

	// Reconfigure the request for proxying
	self.transformHeaders(route, request)
	log.Printf("Origin: %+v\n", origin)
	request.Header.Add("X-Forwarded-Host", request.Host)
	request.Header.Add("X-Forwarded-Proto", request.URL.Scheme)
	request.Header.Add("X-Origin-Host", origin.Host)
	request.Host = origin.Host
	request.URL.Host = origin.Host
	request.URL.Scheme = normalizeScheme(origin.Scheme)
	log.Printf("Request Processing took: %s\n", time.Since(start))

	// Proxy the request
	response, err := transport.RoundTrip(request)
	if err != nil {
		log.Printf("Error in response: %+v\n", err)
		return nil, err //Server is not reachable. Server not working
	}
	log.Printf("Response Time: %s\n", time.Since(start))

	// Reconfigure the response for forwarding to the client
	start = time.Now()
	response.Request.Header = reqHeaders
	self.transformHeaders(route, response)
	log.Printf("Response Rewrite Time: %s\n", time.Since(start))

	return response, nil
}

func (self *Server) Run() int {
	runtime.GOMAXPROCS(int(self.config.Server.Threads))

	router := mux.NewRouter()
	proxy := httputil.ReverseProxy{
		Director:  func(req *http.Request) {},
		Transport: &proxyTransport{self},
	}

	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == `/__portunus_ping__` {
			log.Printf("PING > PONG\n")
			self.handlePingCheck(w, r)
		} else {
			log.Printf("Serving HTTP for request: %+v\n", r)
			proxy.ServeHTTP(w, r)
		}
	})

	self.startup = time.Now()
	log.Printf("Portunus starting on %s", self.config.Server.BindAddress)

	if self.config.Server.TlsEnabled {
		log.Fatal(http.ListenAndServeTLS(
			self.config.Server.BindAddress,
			self.config.Server.TlsCert,
			self.config.Server.TlsKey,
			router,
		))
	} else {
		log.Fatal(http.ListenAndServe(self.config.Server.BindAddress, router))
	}

	return 0
}
