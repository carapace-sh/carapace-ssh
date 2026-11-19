package destination

import "testing"

func FuzzParseDestinationForms(f *testing.F) {
	seeds := []string{
		"",
		"example.com",
		"user@example.com",
		"example.com:22",
		"user@example.com:2222",
		"[2001:db8::1]:22",
		"user@[2001:db8::1]:22",
		"host:/path/to/file",
		"user@host:relative/path",
		"[::1]:/abs/path",
		"sftp://user@host/path",
		"sftp://user@host:22/path",
		"scp://host/",
		"jump1,user@jump2:2222,host3",
		"C:\\local\\path",
		":path",
		"\xff\xfe",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data string) {
		_ = ParseHostPort(data)
		_ = FormatHostPort(ParseHostPort(data).Host, ParseHostPort(data).Port)
		_ = ParseTarget(data)
		_ = FormatTarget(ParseTarget(data))
		_, _, _ = SplitSCPRemote(data)
		_ = JoinSCPRemote("host", data)
		_ = SplitProxyJump(data)
		_ = ParseProxyJump(data)
		_ = JoinProxyJump(ParseProxyJump(data))
		_, _ = ParseURI(data)
		if u, ok := ParseURI(data); ok {
			_ = FormatURI(u)
		}
		_, _ = ParseRemote(data)
		if r, ok := ParseRemote(data); ok {
			_ = FormatRemote(r)
		}
	})
}
