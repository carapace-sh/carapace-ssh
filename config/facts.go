package config

import "sort"

// HostFacts contains best-effort static directives declared for a host.
//
// Values are collected from Host blocks that literally declare the host token
// and from Match blocks that declare host/originalhost criteria for the token.
// Association is static and does not attempt ssh -G style evaluation.
type HostFacts struct {
	HostName              string
	User                  string
	Port                  string
	IdentityFiles         []string
	CertificateFiles      []string
	ProxyJump             string
	ProxyCommand          string
	IdentityAgent         string
	UserKnownHostsFiles   []string
	GlobalKnownHostsFiles []string
}

func cloneFacts(f HostFacts) HostFacts {
	return HostFacts{
		HostName:              f.HostName,
		User:                  f.User,
		Port:                  f.Port,
		IdentityFiles:         append([]string(nil), f.IdentityFiles...),
		CertificateFiles:      append([]string(nil), f.CertificateFiles...),
		ProxyJump:             f.ProxyJump,
		ProxyCommand:          f.ProxyCommand,
		IdentityAgent:         f.IdentityAgent,
		UserKnownHostsFiles:   append([]string(nil), f.UserKnownHostsFiles...),
		GlobalKnownHostsFiles: append([]string(nil), f.GlobalKnownHostsFiles...),
	}
}

func (f *HostFacts) addDirective(keyword string, args []string) {
	if len(args) == 0 {
		return
	}
	value := args[0]
	switch keyword {
	case "hostname":
		f.HostName = value
	case "user":
		f.User = value
	case "port":
		f.Port = value
	case "identityfile":
		f.IdentityFiles = appendUnique(f.IdentityFiles, args...)
	case "certificatefile":
		f.CertificateFiles = appendUnique(f.CertificateFiles, args...)
	case "proxyjump":
		f.ProxyJump = value
	case "proxycommand":
		f.ProxyCommand = stringsJoinArgs(args)
	case "identityagent":
		f.IdentityAgent = value
	case "userknownhostsfile":
		f.UserKnownHostsFiles = appendUnique(f.UserKnownHostsFiles, args...)
	case "globalknownhostsfile":
		f.GlobalKnownHostsFiles = appendUnique(f.GlobalKnownHostsFiles, args...)
	}
}

func appendUnique(dst []string, vals ...string) []string {
	seen := map[string]struct{}{}
	for _, v := range dst {
		seen[v] = struct{}{}
	}
	for _, v := range vals {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		dst = append(dst, v)
	}
	return dst
}

func stringsJoinArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}
	if len(args) == 1 {
		return args[0]
	}
	out := args[0]
	for _, a := range args[1:] {
		out += " " + a
	}
	return out
}

func sortedFactsKeys(m map[string]HostFacts) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
