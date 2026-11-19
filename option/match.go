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
	"all":          {Name: "all", Kind: MatchArgNone},
	"canonical":    {Name: "canonical", Kind: MatchArgNone},
	"final":        {Name: "final", Kind: MatchArgNone},
	"exec":         {Name: "exec", Kind: MatchArgRest},
	"command":      {Name: "command", Kind: MatchArgValue},
	"host":         {Name: "host", Kind: MatchArgValue, ExtractsHosts: true},
	"originalhost": {Name: "originalhost", Kind: MatchArgValue, ExtractsHosts: true},
	"user":         {Name: "user", Kind: MatchArgValue},
	"group":        {Name: "group", Kind: MatchArgValue},
	"localuser":    {Name: "localuser", Kind: MatchArgValue},
	"localaddress": {Name: "localaddress", Kind: MatchArgValue},
	"localnetwork": {Name: "localnetwork", Kind: MatchArgValue},
	"tagged":       {Name: "tagged", Kind: MatchArgValue},
	"address":      {Name: "address", Kind: MatchArgValue},
	"ruser":        {Name: "ruser", Kind: MatchArgValue},
	"rdomain":      {Name: "rdomain", Kind: MatchArgValue},
	"localport":    {Name: "localport", Kind: MatchArgValue},
	"remoteport":   {Name: "remoteport", Kind: MatchArgValue},
	"sessiontype":  {Name: "sessiontype", Kind: MatchArgValue},
	"version":      {Name: "version", Kind: MatchArgValue},
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
