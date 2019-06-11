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
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"

	log "github.com/rabbitt/portunus/portunus/logging"
)

var (
	ErrorNotResolvable = errors.New("Unable to resolve ip")
)

type ProxyTransport struct {
	server *Server
}

func NewProxyTransport(server *Server) *ProxyTransport {
	return &ProxyTransport{server: server}
}

func NewNonChunkedRequest(method, url string, r *http.Request) (newRequest *http.Request, err error) {
	if r.Body == http.NoBody {
		if newRequest, err = http.NewRequest(method, url, r.Body); err != nil {
			return nil, err
		}
		copyHeaders(r.Header, newRequest.Header)
		return
	}

	var buf bytes.Buffer
	if _, err = buf.ReadFrom(r.Body); err != nil {
		return nil, err
	}

	if err = r.Body.Close(); err != nil {
		return nil, err
	}

	newBody := ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
	if newRequest, err = http.NewRequest(method, url, newBody); err != nil {
		return nil, err
	}

	newRequest.ContentLength = int64(buf.Len())
	copyHeaders(r.Header, newRequest.Header)

	return newRequest, nil
}

func notFoundResponse(req *http.Request) *http.Response {
	code := Settings.Response.NotFound.Code
	body := Settings.Response.NotFound.Body
	return &http.Response{
		Status:        fmt.Sprintf("%d %s", code, http.StatusText(code)),
		StatusCode:    code,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBufferString(body)),
		ContentLength: int64(len(body)),
		Request:       req,
		Header:        make(http.Header, 0),
	}
}

func internalServerErrorResponse(req *http.Request) *http.Response {
	code := Settings.Response.ServerError.Code
	body := Settings.Response.ServerError.Body
	return &http.Response{
		Status:        fmt.Sprintf("%d %s", code, http.StatusText(code)),
		StatusCode:    code,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBufferString(body)),
		ContentLength: int64(len(body)),
		Request:       req,
		Header:        make(http.Header, 0),
	}
}

func (pt *ProxyTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	var route *Route
	var origin *url.URL
	var err error
	var ok bool

	// Determine the origin to proxy the request to
	if route, ok = pt.server.router.routeTree.Lookup(request.URL.Path); !ok {
		return notFoundResponse(request), nil
	}

	if origin, err = getUpstream(route, request); err != nil {
		log.Error(err)
		return internalServerErrorResponse(request), nil
	}

	var ips []string

	// verify host is resolvable
	ips, err = net.LookupHost(origin.Hostname())
	if err != nil {
		// When using custom DNS resolvers, DNSError doesn't return
		// the actual custom DNS resolvers, but instead shows the system
		// resolver. While we still don't show which specific resolver
		// failed, at least the code below will show the resolvers used.
		var e *net.DNSError
		switch err.(type) {
		case *net.DNSError:
			if len(Settings.DNS.Resolvers) > 0 {
				e = err.(*net.DNSError)
				e.Server = strings.Join(Settings.DNS.Resolvers, ", or ")
				err = e
			}
		}
		log.ErrorWithFields(err, log.Fields{"origin": origin, "route": route})
		return internalServerErrorResponse(request), nil
	} else if len(ips) <= 0 {
		log.ErrorWithFields(ErrorNotResolvable, log.Fields{"origin": origin, "route": route})
		return internalServerErrorResponse(request), nil
	}

	// keep a copy so it's availble for response rewriting, in case any of the
	// response headers need to include request header details
	var reqHeaders = make(http.Header)
	copyHeaders(request.Header, reqHeaders)

	request.Header.Add("X-Origin-Host", origin.Host)
	request.Header.Add("X-Forwarded-Host", request.Host)
	if request.Header.Get("X-Forwarded-Proto") == "" {
		if Settings.Server.TLS.Enabled {
			request.Header.Add("X-Forwarded-Proto", "https")
		} else {
			request.Header.Add("X-Forwarded-Proto", "http")
		}
	}

	// Allow overriding of the above headers by configuration
	transformHeaders(route, request)

	// Setup for proxying
	request.Host = origin.Host
	request.URL.Host = origin.Host
	request.URL.Scheme = normalizeScheme(origin.Scheme)

	if route.AggregateRequestChunks() {
		switch strings.ToUpper(request.Method) {
		case "POST", "PUT":
			url, _ := url.Parse(fmt.Sprintf("%s%s", origin.String(), request.RequestURI))
			if req, err := NewNonChunkedRequest(request.Method, url.String(), request); err == nil {
				request = req
			} else {
				log.ErrorWithFields("Unable to create request; using original", log.Fields{"error": err})
			}
		}
	}

	// Proxy the request
	log.DebugWithFields("Proxying request", log.Fields{"host": request.Host, "origin": origin})
	TraceEventData(request)

	response, err := pt.server.transport.RoundTrip(request)
	if err != nil {
		log.ErrorWithFields("Upstream responded with Error", log.Fields{"error": err})
		return nil, err //Server is not reachable, or otherwise not working
	}

	TraceEventData(response)

	// Reconfigure the response for forwarding to the client
	response.Request.Header = reqHeaders
	transformHeaders(route, response)

	return response, nil
}
