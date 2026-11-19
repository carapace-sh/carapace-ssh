package destination

import (
	"net/url"
	"strings"
)

// HostPort represents a host and optional port.
type HostPort struct {
	Host string
	Port string
}

// ParseHostPort parses host, [host], host:port, and [host]:port forms.
// It is tolerant and intended for completion helpers, not strict validation.
func ParseHostPort(s string) HostPort {
	s = strings.TrimSpace(s)
	if s == "" {
		return HostPort{}
	}

	if strings.HasPrefix(s, "[") {
		end := strings.Index(s, "]")
		if end > 0 {
			hp := HostPort{Host: s[1:end]}
			rest := s[end+1:]
			if port, ok := strings.CutPrefix(rest, ":"); ok {
				hp.Port = port
			}
			return hp
		}
	}

	// Unbracketed IPv6 or ambiguous forms: only treat last :port when there is
	// exactly one colon to avoid breaking raw IPv6 literals.
	if strings.Count(s, ":") == 1 {
		host, port, ok := strings.Cut(s, ":")
		if ok {
			return HostPort{Host: host, Port: port}
		}
	}

	return HostPort{Host: s}
}

// FormatHostPort formats host and optional port, bracketing IPv6 hosts when a
// port is present.
func FormatHostPort(host, port string) string {
	host = strings.TrimSpace(host)
	port = strings.TrimSpace(port)
	if host == "" {
		return ""
	}
	needsBracket := strings.Contains(host, ":") && !strings.HasPrefix(host, "[")
	if port == "" {
		if needsBracket {
			return "[" + host + "]"
		}
		return host
	}
	if needsBracket || strings.HasPrefix(host, "[") {
		if !strings.HasPrefix(host, "[") {
			host = "[" + host + "]"
		}
		return host + ":" + port
	}
	return host + ":" + port
}

// Target represents a user@host destination with optional port.
type Target struct {
	User string
	Host string
	Port string
}

// ParseTarget parses [user@]host, [user@]host:port, and [user@][host]:port
// forms. It does not parse scp path separators.
func ParseTarget(s string) Target {
	s = strings.TrimSpace(s)
	if s == "" {
		return Target{}
	}

	user := ""
	hostpart := s
	if at := strings.LastIndex(s, "@"); at >= 0 {
		user = s[:at]
		hostpart = s[at+1:]
	}

	hp := ParseHostPort(hostpart)
	return Target{User: user, Host: hp.Host, Port: hp.Port}
}

// FormatTarget formats a user@host[:port] target.
func FormatTarget(t Target) string {
	host := FormatHostPort(t.Host, t.Port)
	if host == "" {
		return strings.TrimSpace(t.User)
	}
	if strings.TrimSpace(t.User) == "" {
		return host
	}
	return strings.TrimSpace(t.User) + "@" + host
}

// SplitSCPRemote splits an scp/sftp-style remote argument into target and path
// components. It returns ok=false for local paths without a remote host
// separator. Windows drive-letter paths (e.g. C:\path) are treated as local.
func SplitSCPRemote(s string) (target string, path string, ok bool) {
	if s == "" {
		return "", "", false
	}
	if isWindowsDrivePath(s) {
		return "", s, false
	}

	idx := scpSeparatorIndex(s)
	if idx < 0 {
		return "", s, false
	}
	if idx == 0 {
		return "", s[idx+1:], true
	}
	return s[:idx], s[idx+1:], true
}

// JoinSCPRemote joins a remote target and path into an scp/sftp-style argument.
// If target is empty, path is returned unchanged.
func JoinSCPRemote(target, path string) string {
	if target == "" {
		return path
	}
	return target + ":" + path
}

func scpSeparatorIndex(s string) int {
	inBracket := false
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '[':
			inBracket = true
		case ']':
			if inBracket {
				inBracket = false
			}
		case ':':
			if !inBracket {
				return i
			}
		}
	}
	return -1
}

func isWindowsDrivePath(s string) bool {
	if len(s) < 2 {
		return false
	}
	c := s[0]
	if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
		return false
	}
	return s[1] == ':'
}

// SplitProxyJump splits a ProxyJump specification into hop strings.
func SplitProxyJump(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	hops := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		hops = append(hops, part)
	}
	return hops
}

// ParseProxyJump parses a ProxyJump specification into hop targets.
func ParseProxyJump(s string) []Target {
	parts := SplitProxyJump(s)
	if len(parts) == 0 {
		return []Target{}
	}
	targets := make([]Target, 0, len(parts))
	for _, part := range parts {
		targets = append(targets, ParseTarget(part))
	}
	return targets
}

// JoinProxyJump joins hop targets into a ProxyJump specification.
func JoinProxyJump(hops []Target) string {
	formatted := make([]string, 0, len(hops))
	for _, hop := range hops {
		s := FormatTarget(hop)
		if s == "" {
			continue
		}
		formatted = append(formatted, s)
	}
	return strings.Join(formatted, ",")
}

// URI represents a remote destination URI (e.g. sftp://user@host:port/path).
type URI struct {
	Scheme   string
	User     string
	Host     string
	Port     string
	Path     string
	RawQuery string
	Fragment string
}

// ParseURI parses an scp/sftp-style destination URI. It returns ok=false for
// strings without a recognized URI scheme separator.
func ParseURI(s string) (URI, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return URI{}, false
	}
	idx := strings.Index(s, "://")
	if idx <= 0 {
		return URI{}, false
	}
	scheme := strings.ToLower(s[:idx])
	u, err := url.Parse(s)
	if err != nil || u == nil || u.Host == "" {
		return URI{}, false
	}
	parsed := URI{
		Scheme:   scheme,
		Host:     u.Hostname(),
		Port:     u.Port(),
		Path:     u.EscapedPath(),
		RawQuery: u.RawQuery,
		Fragment: u.Fragment,
	}
	if parsed.Path == "" {
		parsed.Path = u.Path
	}
	if u.User != nil {
		parsed.User = u.User.Username()
	}
	return parsed, true
}

// FormatURI formats a destination URI. It returns an empty string when scheme or
// host is missing.
func FormatURI(u URI) string {
	scheme := strings.ToLower(strings.TrimSpace(u.Scheme))
	host := strings.TrimSpace(u.Host)
	if scheme == "" || host == "" {
		return ""
	}
	out := scheme + "://"
	if user := strings.TrimSpace(u.User); user != "" {
		out += user + "@"
	}
	out += FormatHostPort(host, u.Port)
	path := u.Path
	if path != "" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	out += path
	if q := strings.TrimPrefix(u.RawQuery, "?"); q != "" {
		out += "?" + q
	}
	if f := strings.TrimPrefix(u.Fragment, "#"); f != "" {
		out += "#" + f
	}
	return out
}

// Remote represents a parsed remote destination in URI or scp-style form.
type Remote struct {
	Scheme   string
	Target   Target
	Path     string
	RawQuery string
	Fragment string
	URI      bool
}

// ParseRemote parses a remote destination from either URI or scp-style syntax.
// It returns ok=false for local paths without a remote host separator.
func ParseRemote(s string) (Remote, bool) {
	if uri, ok := ParseURI(s); ok {
		return Remote{
			Scheme:   uri.Scheme,
			Target:   Target{User: uri.User, Host: uri.Host, Port: uri.Port},
			Path:     uri.Path,
			RawQuery: uri.RawQuery,
			Fragment: uri.Fragment,
			URI:      true,
		}, true
	}

	target, path, ok := SplitSCPRemote(s)
	if !ok {
		return Remote{}, false
	}
	return Remote{
		Target: ParseTarget(target),
		Path:   path,
	}, true
}

// FormatRemote formats a remote destination. URI remotes are formatted as URIs;
// scp-style remotes are formatted as target:path.
func FormatRemote(r Remote) string {
	if r.URI {
		return FormatURI(URI{
			Scheme:   r.Scheme,
			User:     r.Target.User,
			Host:     r.Target.Host,
			Port:     r.Target.Port,
			Path:     r.Path,
			RawQuery: r.RawQuery,
			Fragment: r.Fragment,
		})
	}
	return FormatTarget(r.Target) + ":" + r.Path
}

// IsURI reports whether s parses as a destination URI.
func IsURI(s string) bool {
	_, ok := ParseURI(s)
	return ok
}

// IsRemote reports whether s parses as a remote destination (URI or scp-style).
func IsRemote(s string) bool {
	_, ok := ParseRemote(s)
	return ok
}
