package portunus

import (
	"fmt"
	// "log"
	"net/http"
	"strings"
)

func interpolate(value string, route *Route, httpObj interface{}) string {
	var ok bool
	var req *http.Request
	var resp *http.Response

	if req, ok = httpObj.(*http.Request); !ok {
		resp = httpObj.(*http.Response)
		req = resp.Request
	}

	if strings.Contains(value, `{{route.`) {
		value = strings.Replace(value, `{{route.name}}`, route.name, -1)
		value = strings.Replace(value, `{{route.match}}`, route.matchedPath, -1)
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
