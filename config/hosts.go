package config

import (
	"sort"
	"strings"
)

func normalizeHostToken(token string) string {
	return strings.TrimSpace(token)
}

func isCompletableHost(token string) bool {
	token = normalizeHostToken(token)
	if token == "" || strings.HasPrefix(token, "!") {
		return false
	}
	if strings.ContainsAny(token, "*?[]{}()|+^$! \t\r\n") {
		return false
	}
	return true
}

func sortedUniqueHosts(hosts map[string]struct{}) []string {
	if len(hosts) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(hosts))
	for h := range hosts {
		out = append(out, h)
	}
	sort.Strings(out)
	return out
}
