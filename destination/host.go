package destination

import "strings"

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
			if strings.HasPrefix(rest, ":") {
				hp.Port = strings.TrimPrefix(rest, ":")
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
