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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	log "github.com/rabbitt/portunus/portunus/logging"
	config "github.com/spf13/viper"
)

type proxyTransport struct {
	server *Server
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
	} else {
		newRequest.ContentLength = int64(buf.Len())
		copyHeaders(r.Header, newRequest.Header)
	}

	return newRequest, nil
}

func notFoundResponse(req *http.Request) *http.Response {
	code := config.GetInt("error_pages.not_found.code")
	body := config.GetString("error_pages.not_found.body")
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
	code := config.GetInt("error_pages.server_error.code")
	body := config.GetString("error_pages.server_error.body")
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

func (self *proxyTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	var route *Route
	var origin *url.URL
	var err error
	var ok bool

	// Determine the origin to proxy the request to
	if route, ok = self.server.router.routeTree.Lookup(request.URL.Path); !ok {
		return notFoundResponse(request), nil
	}

	if origin, err = url.Parse(getOrigin(route, request)); err != nil {
		log.Error(err)
		return internalServerErrorResponse(request), nil
	}

	// keep a copy so it's availble for response rewriting, in case any of the
	// response headers need to include request header details
	var reqHeaders = make(http.Header)
	copyHeaders(request.Header, reqHeaders)

	request.Header.Add("X-Origin-Host", origin.Host)
	request.Header.Add("X-Forwarded-Host", request.Host)
	if request.Header.Get("X-Forwarded-Proto") == "" {
		if config.GetBool("server.tls.enabled") {
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

	response, err := self.server.transport.RoundTrip(request)
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
