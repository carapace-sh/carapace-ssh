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

		negated := strings.HasPrefix(critToken, "!")
		crit, known := option.LookupMatchCriterion(critToken)
		if known {
			switch {
			case crit.ExtractsHosts:
				value, next := matchCriterionValue(args, i, inlineValue, hasInline)
				i = next
				if negated || value == "" {
					continue
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
				_, next := matchCriterionValue(args, i, inlineValue, hasInline)
				i = next
			}
			continue
		}

		// Unknown criterion: conservatively consume one argument if present to
		// preserve alignment for later host/originalhost where possible, unless it
		// looks like another criterion token. This is best-effort.
		if hasInline {
			if inlineValue == "" && i+1 < len(args) {
				next := strings.TrimSpace(args[i+1])
				if next != "" && !isKnownCriterionToken(next) && !strings.HasPrefix(next, "=") {
					i++
				}
			}
			continue
		}
		if i+1 < len(args) {
			next := strings.TrimSpace(args[i+1])
			if next == "=" {
				if i+2 < len(args) && !isKnownCriterionToken(args[i+2]) {
					i += 2
				} else {
					i++
				}
				continue
			}
			if strings.HasPrefix(next, "=") {
				// "=value" form; treat as consumed argument.
				i++
				continue
			}
			if !isKnownCriterionToken(next) {
				i++
			}
		}
	}
	return candidates
}

func matchCriterionValue(args []string, i int, inlineValue string, hasInline bool) (string, int) {
	if hasInline {
		if inlineValue != "" {
			return inlineValue, i
		}
		if i+1 < len(args) {
			next := strings.TrimSpace(args[i+1])
			if next == "" || strings.HasPrefix(next, "=") || isKnownCriterionToken(next) {
				// Empty inline value: do not swallow the next criterion token.
				return "", i
			}
			return args[i+1], i + 1
		}
		return "", i
	}
	if i+1 >= len(args) {
		return "", i
	}
	next := strings.TrimSpace(args[i+1])
	if next == "=" {
		if i+2 < len(args) {
			value := strings.TrimSpace(args[i+2])
			if value == "" || strings.HasPrefix(value, "=") || isKnownCriterionToken(value) {
				// Spaced equals with a following criterion token: consume only the "=".
				return "", i + 1
			}
			return args[i+2], i + 2
		}
		return "", i + 1
	}
	if strings.HasPrefix(next, "=") {
		return strings.TrimPrefix(next, "="), i + 1
	}
	return args[i+1], i + 1
}

func isKnownCriterionToken(s string) bool {
	return option.IsKnownMatchCriterion(s)
}
