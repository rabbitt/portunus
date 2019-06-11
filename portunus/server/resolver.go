package server

import (
	"net"

	"github.com/benburkert/dns"
	log "github.com/rabbitt/portunus/portunus/logging"
)

type DNSError net.DNSError

var DefaultResolver *net.Resolver = net.DefaultResolver

func NewResolver(servers []string) *net.Resolver {
	var nameservers dns.NameServers
	for _, server := range servers {
		nameservers = append(nameservers, &net.UDPAddr{
			IP: net.ParseIP(server), Port: 53,
		})
	}

	return &net.Resolver{
		PreferGo: true,

		Dial: (&dns.Client{
			Resolver: new(dns.Cache),
			Transport: &dns.Transport{
				Proxy: nameservers.RoundRobin(),
			},
		}).Dial,
	}
}

func SetResolvers(nameservers []string) {
	if len(nameservers) <= 0 {
		net.DefaultResolver = DefaultResolver
		log.InfoWithFields("DNS Resolvers", log.Fields{"resolvers": "<system>"})
	} else {
		net.DefaultResolver = NewResolver(nameservers)
		log.InfoWithFields("DNS Resolvers", log.Fields{"resolvers": nameservers})
	}
}
