package knownhosts

import (
	"bufio"
	"io"
	"strings"
)

func parseReader(r io.Reader, add func(string)) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		parseLine(scanner.Text(), add)
	}
	return scanner.Err()
}

func parseLine(line string, add func(string)) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return
	}

	fields := strings.Fields(line)
	if len(fields) == 0 {
		return
	}

	marker := ""
	// Optional markers @cert-authority / @revoked
	switch fields[0] {
	case "@cert-authority", "@revoked":
		marker = fields[0]
		if len(fields) < 2 {
			return
		}
		fields = fields[1:]
	}

	hostField := fields[0]
	// known_hosts lines require host field plus key type and key material.
	if hostField == "" || len(fields) < 3 {
		return
	}

	// Revoked host keys must not be accepted; do not suggest them for completion.
	if marker == "@revoked" {
		return
	}

	for _, token := range strings.Split(hostField, ",") {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		// Hashed hosts cannot be reversed for completion.
		if strings.HasPrefix(token, "|") {
			continue
		}

		host := token
		// [host]:port form
		if strings.HasPrefix(token, "[") {
			end := strings.Index(token, "]")
			if end <= 1 {
				continue
			}
			host = token[1:end]
			// Keep bracketed host token parsing tolerant; ignore malformed suffixes
			// beyond the closing bracket except for standard :port presence.
		}

		if isCompletableHost(host) {
			add(host)
		}
	}
}
