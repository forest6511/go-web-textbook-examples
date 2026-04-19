package secret

import "log/slog"

type Password string

func (p Password) LogValue() slog.Value {
	return slog.StringValue("REDACTED")
}

type Token string

func (t Token) LogValue() slog.Value {
	if len(t) <= 4 {
		return slog.StringValue("****")
	}
	return slog.StringValue("****" + string(t[len(t)-4:]))
}
