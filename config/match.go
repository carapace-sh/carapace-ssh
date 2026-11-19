package config

import (
	"strings"

	"github.com/carapace-sh/carapace-ssh/option"
)

// extractMatchHosts returns host candidates declared via Match host/originalhost.
// It is tolerant: unknown criteria do not abort extraction, and exec/command are
// never executed (their arguments are treated as opaque text to skip when possible).
func extractMatchHosts(args []string) []string {
	candidates := make([]string, 0)
	for i := 0; i < len(args); i++ {
		token := strings.TrimSpace(args[i])
		if token == "" {
			continue
		}

		critToken := token
		inlineValue := ""
		hasInline := false
		if idx := strings.Index(token, "="); idx >= 0 {
			critToken = token[:idx]
			inlineValue = token[idx+1:]
			hasInline = true
		}

		crit, known := option.LookupMatchCriterion(critToken)
		if known {
			switch {
			case crit.ExtractsHosts:
				value := inlineValue
				if !hasInline {
					if i+1 >= len(args) {
						continue
					}
					value = args[i+1]
					i++
				}
				for _, v := range splitCommaList(value) {
					candidates = append(candidates, v)
				}
			case crit.Kind == option.MatchArgRest:
				// Command text is typically rest-of-line; never execute. Stop to avoid
				// misinterpreting command text as further criteria.
				return candidates
			case crit.Kind == option.MatchArgNone:
				// zero-arg criterion
			case crit.Kind == option.MatchArgValue:
				if !hasInline && i+1 < len(args) {
					i++
				}
			}
			continue
		}

		// Unknown criterion: conservatively consume one argument if present to
		// preserve alignment for later host/originalhost where possible, unless it
		// looks like another criterion token. This is best-effort.
		if hasInline {
			continue
		}
		if i+1 < len(args) {
			next := strings.ToLower(strings.TrimSpace(args[i+1]))
			nextBase := strings.TrimPrefix(next, "!")
			if eq := strings.Index(nextBase, "="); eq >= 0 {
				nextBase = nextBase[:eq]
			}
			if !isKnownCriterionToken(nextBase) {
				i++
			}
		}
	}
	return candidates
}

func isKnownCriterionToken(s string) bool {
	if option.IsKnownMatchCriterion(s) {
		return true
	}
	_, ok := option.LookupMatchCriterion(s)
	return ok
}
