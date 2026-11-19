package option

import (
	"sort"
	"strings"
)

// Keyword describes basic metadata for an ssh_config keyword.
type Keyword struct {
	Name string
}

// NormalizeKeyword returns a lowercased keyword name for lookup/comparison.
func NormalizeKeyword(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

var keywords = map[string]Keyword{
	"host":                 {Name: "Host"},
	"match":                {Name: "Match"},
	"include":              {Name: "Include"},
	"hostname":             {Name: "HostName"},
	"user":                 {Name: "User"},
	"port":                 {Name: "Port"},
	"identityfile":         {Name: "IdentityFile"},
	"certificatefile":      {Name: "CertificateFile"},
	"identityagent":        {Name: "IdentityAgent"},
	"proxyjump":            {Name: "ProxyJump"},
	"proxycommand":         {Name: "ProxyCommand"},
	"userknownhostsfile":   {Name: "UserKnownHostsFile"},
	"globalknownhostsfile": {Name: "GlobalKnownHostsFile"},
}

// LookupKeyword returns metadata for a known ssh_config keyword.
func LookupKeyword(name string) (Keyword, bool) {
	k, ok := keywords[NormalizeKeyword(name)]
	return k, ok
}

// IsKnownKeyword reports whether the keyword is known to this metadata set.
func IsKnownKeyword(name string) bool {
	_, ok := LookupKeyword(name)
	return ok
}

// Keywords returns known keywords sorted by canonical name.
func Keywords() []Keyword {
	out := make([]Keyword, 0, len(keywords))
	for _, k := range keywords {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
