package option

import (
	"sort"
	"strings"
)

// MatchArgKind describes how a Match criterion consumes arguments.
type MatchArgKind int

const (
	MatchArgNone MatchArgKind = iota
	MatchArgValue
	MatchArgRest
)

// MatchCriterion describes parsing metadata for a Match criterion.
type MatchCriterion struct {
	Name          string
	Kind          MatchArgKind
	ExtractsHosts bool
}

var matchCriteria = map[string]MatchCriterion{
	// ssh_config(5) Match criteria (OpenSSH 10.6p1 man page).
	"all":          {Name: "all", Kind: MatchArgNone},
	"canonical":    {Name: "canonical", Kind: MatchArgNone},
	"final":        {Name: "final", Kind: MatchArgNone},
	"exec":         {Name: "exec", Kind: MatchArgRest},
	"command":      {Name: "command", Kind: MatchArgValue},
	"host":         {Name: "host", Kind: MatchArgValue, ExtractsHosts: true},
	"originalhost": {Name: "originalhost", Kind: MatchArgValue, ExtractsHosts: true},
	"user":         {Name: "user", Kind: MatchArgValue},
	"localuser":    {Name: "localuser", Kind: MatchArgValue},
	"localnetwork": {Name: "localnetwork", Kind: MatchArgValue},
	"tagged":       {Name: "tagged", Kind: MatchArgValue},
	"sessiontype":  {Name: "sessiontype", Kind: MatchArgValue},
	"version":      {Name: "version", Kind: MatchArgValue},

	// Tolerant extra criteria seen in other OpenSSH contexts/variants or older
	// syntax; kept as value args for safe alignment. Not all are documented in
	// ssh_config(5) Match for this OpenSSH version.
	"address":      {Name: "address", Kind: MatchArgValue},
	"group":        {Name: "group", Kind: MatchArgValue},
	"localgroup":   {Name: "localgroup", Kind: MatchArgValue},
	"localaddress": {Name: "localaddress", Kind: MatchArgValue},
	"localport":    {Name: "localport", Kind: MatchArgValue},
	"rdomain":      {Name: "rdomain", Kind: MatchArgValue},
	"remoteport":   {Name: "remoteport", Kind: MatchArgValue},
	"ruser":        {Name: "ruser", Kind: MatchArgValue},
}

// LookupMatchCriterion returns metadata for a Match criterion.
func LookupMatchCriterion(name string) (MatchCriterion, bool) {
	name = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(name, "!")))
	if idx := strings.Index(name, "="); idx >= 0 {
		name = name[:idx]
	}
	c, ok := matchCriteria[name]
	return c, ok
}

// IsKnownMatchCriterion reports whether the criterion name is known.
func IsKnownMatchCriterion(name string) bool {
	_, ok := LookupMatchCriterion(name)
	return ok
}

// IsZeroArgMatchCriterion reports whether the criterion consumes no arguments.
func IsZeroArgMatchCriterion(name string) bool {
	c, ok := LookupMatchCriterion(name)
	return ok && c.Kind == MatchArgNone
}

// MatchCriteria returns known Match criteria sorted by name.
func MatchCriteria() []MatchCriterion {
	out := make([]MatchCriterion, 0, len(matchCriteria))
	for _, c := range matchCriteria {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
