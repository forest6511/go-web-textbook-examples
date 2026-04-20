package observability

import (
	"testing"

	"github.com/getsentry/sentry-go"
)

func TestScrubPII_RemovesUserEmailAndIP(t *testing.T) {
	ev := &sentry.Event{
		User: sentry.User{
			ID:        "42",
			Email:     "user@example.com",
			IPAddress: "203.0.113.1",
		},
	}
	got := ScrubPII(ev, nil)
	if got.User.Email != "" {
		t.Errorf("Email = %q, want empty", got.User.Email)
	}
	if got.User.IPAddress != "" {
		t.Errorf("IPAddress = %q, want empty", got.User.IPAddress)
	}
	if got.User.ID != "42" {
		t.Errorf("User.ID = %q, want to preserve for error tracking", got.User.ID)
	}
}

func TestScrubPII_RemovesAuthHeadersAndCookies(t *testing.T) {
	ev := &sentry.Event{
		Request: &sentry.Request{
			Cookies: "session=abc; refresh_token=xyz",
			Headers: map[string]string{
				"Authorization": "Bearer secret-token",
				"Cookie":        "session=abc",
				"User-Agent":    "curl/8.0",
			},
		},
	}
	got := ScrubPII(ev, nil)
	if got.Request.Cookies != "" {
		t.Errorf("Cookies = %q, want empty", got.Request.Cookies)
	}
	if _, ok := got.Request.Headers["Authorization"]; ok {
		t.Error("Authorization header should be removed")
	}
	if _, ok := got.Request.Headers["Cookie"]; ok {
		t.Error("Cookie header should be removed")
	}
	if got.Request.Headers["User-Agent"] != "curl/8.0" {
		t.Error("User-Agent should be preserved")
	}
}

func TestScrubPII_RedactsAuthBody(t *testing.T) {
	ev := &sentry.Event{
		Request: &sentry.Request{
			URL:  "/api/v1/auth/login",
			Data: `{"email":"u@example.com","password":"p@ssw0rd"}`,
		},
	}
	got := ScrubPII(ev, nil)
	if got.Request.Data != "[REDACTED]" {
		t.Errorf("Data = %q, want [REDACTED]", got.Request.Data)
	}
}

func TestScrubPII_RedactsBodyContainingPassword(t *testing.T) {
	ev := &sentry.Event{
		Request: &sentry.Request{
			URL:  "/api/v1/tasks",
			Data: `{"title":"t","password":"accidental-leak"}`,
		},
	}
	got := ScrubPII(ev, nil)
	if got.Request.Data != "[REDACTED]" {
		t.Errorf("Data = %q, want [REDACTED] even outside /auth/*", got.Request.Data)
	}
}

func TestScrubPII_KeepsNonSensitiveBody(t *testing.T) {
	ev := &sentry.Event{
		Request: &sentry.Request{
			URL:  "/api/v1/tasks",
			Data: `{"title":"牛乳を買う"}`,
		},
	}
	got := ScrubPII(ev, nil)
	if got.Request.Data != `{"title":"牛乳を買う"}` {
		t.Errorf("non-sensitive body was redacted: %q", got.Request.Data)
	}
}

func TestScrubPII_NilEvent(t *testing.T) {
	if got := ScrubPII(nil, nil); got != nil {
		t.Errorf("nil input should return nil, got %#v", got)
	}
}
