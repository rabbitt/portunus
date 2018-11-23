package portunus

const (
	ConfigName = "portunus"
)

var (
	ConfigSearchPaths = []string{
		"/etc/portunus",
		"$HOME/.portunus",
		".",
	}
)
