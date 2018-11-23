package portunus

import (
	"bytes"
	"log"
	"strings"
	"sync"

	"github.com/fanyang01/radix"
)

type Route struct {
	name        string
	matchedPath string
}

type RouteTree struct {
	mutex sync.Mutex
	radix *radix.PatternTrie
}

func NewRouteTree() *RouteTree {
	return &RouteTree{}
}

func (self *RouteTree) Load(routes map[string]ConfigRoute) *RouteTree {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	self.radix = radix.NewPatternTrie()

	for name, route := range routes {
		for _, path := range route.Paths {
			log.Printf("Adding route: [%s] routes to: [%s]\n", path, name)
			self.radix.Add(normalizePath(path), &Route{name: name, matchedPath: path})
		}
	}

	return self
}

func normalizePath(path string) string {
	if !strings.HasPrefix(path, "/") {
		buffer := bytes.NewBufferString(`/`)
		buffer.WriteString(path)
		return buffer.String()
	}

	return path
}

func (self *RouteTree) insert(s string, route *Route) (*Route, bool) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if v, has := self.radix.Add(normalizePath(s), route); has {
		oldRoute := v.(*Route)
		return oldRoute, true
	}
	return nil, false
}

func (self *RouteTree) Lookup(s string) (*Route, bool) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if v, ok := self.radix.Lookup(s); ok {
		return v.(*Route), true
	}

	return nil, false
}
