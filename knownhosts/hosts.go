package knownhosts

import "sort"

func isCompletableHost(host string) bool {
	if host == "" {
		return false
	}
	// Exclude wildcards, negations, markers, and whitespace/control-like separators.
	for _, r := range host {
		switch r {
		case '*', '?', '!', '[', ']', '{', '}', '(', ')', '|', '+', '^', '$', '@', ' ', '\t', '\r', '\n', ',':
			return false
		}
	}
	return true
}

func sortedUnique(hosts map[string]struct{}) []string {
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

// Merge returns a deduplicated, sorted union of host candidate slices.
func Merge(slices ...[]string) []string {
	set := map[string]struct{}{}
	for _, slice := range slices {
		for _, host := range slice {
			if host == "" {
				continue
			}
			set[host] = struct{}{}
		}
	}
	return sortedUnique(set)
}
