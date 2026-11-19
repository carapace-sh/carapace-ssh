package config

import "strings"

var zeroArgMatchCriteria = map[string]struct{}{
	"all":       {},
	"canonical": {},
	"final":     {},
}

// extractMatchHosts returns host candidates declared via Match host/originalhost.
// It is tolerant: unknown criteria do not abort extraction, and exec/command are
// never executed (their arguments are treated as opaque text to skip when possible).
func extractMatchHosts(args []string) []string {
	candidates := make([]string, 0)
	for i := 0; i < len(args); i++ {
		crit := strings.ToLower(strings.TrimSpace(args[i]))
		if crit == "" {
			continue
		}
		negated := strings.HasPrefix(crit, "!")
		if negated {
			crit = strings.TrimPrefix(crit, "!")
		}

		switch crit {
		case "host", "originalhost":
			if i+1 >= len(args) {
				continue
			}
			for _, v := range splitCommaList(args[i+1]) {
				candidates = append(candidates, v)
			}
			i++
		case "exec", "command":
			// Command text is typically rest-of-line; never execute. Stop to avoid
			// misinterpreting command text as further criteria.
			return candidates
		case "all", "canonical", "final":
			// zero-arg criteria
		case "user", "localuser", "localnetwork", "tagged", "address", "ruser", "rdomain", "localport", "remoteport":
			if i+1 < len(args) {
				i++
			}
		default:
			// Unknown criterion: conservatively consume one argument if present to
			// preserve alignment for later host/originalhost where possible, unless it
			// looks like another criterion keyword token. This is best-effort.
			if i+1 < len(args) {
				next := strings.ToLower(strings.TrimSpace(args[i+1]))
				nextBase := strings.TrimPrefix(next, "!")
				if _, isZero := zeroArgMatchCriteria[nextBase]; !isZero && !looksLikeMatchCriterion(nextBase) {
					i++
				}
			}
			_ = negated
		}
	}
	return candidates
}

func looksLikeMatchCriterion(s string) bool {
	switch s {
	case "host", "originalhost", "exec", "command", "all", "canonical", "final",
		"user", "localuser", "localnetwork", "tagged", "address", "ruser", "rdomain", "localport", "remoteport":
		return true
	default:
		return false
	}
}
