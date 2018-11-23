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
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	log "github.com/rabbitt/portunus/portunus/logging"
	config "github.com/spf13/viper"
)

func transformHeaders(route *Route, httpObj interface{}) {
	var transforms *config.Viper
	var headers *http.Header

	if req, ok := httpObj.(*http.Request); ok {
		headers = &req.Header
		transforms = config.Sub("transform.request")
		log.DebugWithFields("Rewriting Request headers", log.Fields{"route": route, "request.headers": headers})
	} else {
		headers = &(httpObj.(*http.Response)).Header
		transforms = config.Sub("transform.response")
		log.DebugWithFields("Rewriting Response headers", log.Fields{"route": route, "response.headers": headers})
	}

	if transforms != nil {
		for header, value := range transforms.GetStringMapString("insert") {
			headers.Add(header, interpolate(value, route, httpObj))
			log.DebugWithFields("Adding Header", log.Fields{"header": header, "value.new": headers.Get(header), "value.old": value})
		}

		for header, value := range transforms.GetStringMapString("override") {
			headers.Set(header, interpolate(value, route, httpObj))
			log.DebugWithFields("Overwriting Header", log.Fields{"header": header, "value.new": headers.Get(header), "value.old": value})
		}

		for _, header := range transforms.GetStringSlice("delete") {
			log.DebugWithFields("Deleting Header", log.Fields{"header": header, "value.old": headers.Get(header)})
			headers.Del(header)
		}
	}
}

func getOrigin(route *Route, req *http.Request) string {
	// TODO: perform a DNS lookup of the dynamic origin to ensure it's resolvable.
	// If it's not, return the Unmatched Host and log an error.
	// Note: DNS lookups might require a caching mechanism
	if route != nil {
		return interpolate(route.Upstream, route, req)
	} else {
		return interpolate("", &Route{Name: "unmatched", MatchedPath: "<unmatched>"}, req)
	}
}

func debug(data []byte, err error) {
	if err == nil {
		fmt.Printf("%s\n\n", data)
	} else {
		log.Fatalf("%s\n\n", err)
	}
}

func copyHeaders(src, dst http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func RequestHeaderTracingEnabled(r *http.Request) bool {
	return RequestBodyTracingEnabled(r) || (r.Header.Get("X-Trace-Request-Headers") == "true")
}

func RequestBodyTracingEnabled(r *http.Request) bool {
	return (r.Header.Get("X-Trace-Request-Body") == "true")
}

func ResponseHeaderTracingEnabled(r *http.Response) bool {
	return ResponseBodyTracingEnabled(r) || (r.Request.Header.Get("X-Trace-Response-Headers") == "true")
}

func ResponseBodyTracingEnabled(r *http.Response) bool {
	return (r.Request.Header.Get("X-Trace-Response-Body") == "true")
}

func TraceEventData(r interface{}) {
	if request, ok := r.(*http.Request); ok {
		log.InfoWithFields("Request Tracing", log.Fields{
			"header.tracing": RequestHeaderTracingEnabled(request),
			"body.tracing":   RequestBodyTracingEnabled(request),
		})
		if RequestHeaderTracingEnabled(request) {
			debug(httputil.DumpRequest(request, RequestBodyTracingEnabled(request)))
		}
	} else {
		response := r.(*http.Response)
		log.InfoWithFields("Response Tracing", log.Fields{
			"header.tracing": ResponseHeaderTracingEnabled(response),
			"body.tracing":   ResponseBodyTracingEnabled(response),
		})
		if ResponseHeaderTracingEnabled(response) {
			debug(httputil.DumpResponse(response, ResponseBodyTracingEnabled(response)))
		}
	}
}

func normalizeScheme(scheme string) string {
	if scheme == "https" {
		return "https"
	}
	return "http"
}

func interpolate(value string, route *Route, httpObj interface{}) string {
	var ok bool
	var req *http.Request
	var resp *http.Response

	if req, ok = httpObj.(*http.Request); !ok {
		resp = httpObj.(*http.Response)
		req = resp.Request
	}

	if strings.Contains(value, `{{route.`) {
		value = strings.Replace(value, `{{route.name}}`, route.Name, -1)
		value = strings.Replace(value, `{{route.match}}`, route.MatchedPath, -1)
	}

	if req != nil {
		for header, values := range req.Header {
			needle := fmt.Sprintf("{{req.header.%s}}", strings.ToLower(header))
			if strings.Contains(value, needle) {
				value = strings.Replace(value, needle, strings.Join(values, ``), -1)
			}
		}
	}

	if resp != nil {
		for header, values := range resp.Header {
			needle := fmt.Sprintf("{{res.header.%s}}", strings.ToLower(header))
			if strings.Contains(value, needle) {
				value = strings.Replace(value, needle, strings.Join(values, ``), -1)
			}
		}
	}

	return value
}
