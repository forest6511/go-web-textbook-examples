package observability

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
)

// InitSentry は SENTRY_DSN があれば Sentry SDK を初期化する。
// 本番では必ず DSN をセットする想定。dev/test では未設定で no-op。
// Release と Environment は Ch 11 の GitHub Actions で埋める。
func InitSentry(release, environment string) error {
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		return nil // dev/test は Sentry を使わない
	}
	return sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Release:          release,
		Environment:      environment,
		SampleRate:       1.0, // error は全件送信（課金対象ではない）
		TracesSampleRate: 0.0, // Tracing は OTel 側で行うため Sentry は error だけ拾う
		AttachStacktrace: true,
		MaxBreadcrumbs:   100,
		BeforeSend:       ScrubPII,
	})
}

// FlushSentry は shutdown 時に pending events を送信完了まで待つ。
// graceful shutdown の defer 順序で observability 系は最後に閉じる。
func FlushSentry(timeout time.Duration) {
	sentry.Flush(timeout)
}

// ScrubPII はイベント送信前に PII を削除する。
// Sentry へ漏れやすい典型パスを塞ぐ（Ch 12 本文参照）:
//   - User の Email
//   - Request.Cookies（refresh_token が入る）
//   - Request.Headers["Authorization"]
//   - Request.Data 内の password フィールド
//
// 追加項目はサービス固有のフィールド名で BeforeSend を拡張する。
func ScrubPII(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
	if event == nil {
		return nil
	}
	// User 情報から Email / IP を削除（ID はエラー追跡のため残す）
	event.User.Email = ""
	event.User.IPAddress = ""

	if event.Request != nil {
		event.Request.Cookies = ""
		if event.Request.Headers != nil {
			delete(event.Request.Headers, "Authorization")
			delete(event.Request.Headers, "Cookie")
		}
		// Data は raw body の文字列。password / refresh_token を含む場合は
		// 個別マスクが難しいため /auth/* 配下ではまるごと捨てる。
		if strings.HasPrefix(event.Request.URL, "/api/v1/auth/") ||
			strings.Contains(event.Request.Data, "\"password\"") ||
			strings.Contains(event.Request.Data, "\"refresh_token\"") {
			event.Request.Data = "[REDACTED]"
		}
	}
	return event
}

// BuildInfo はビルド時にリンカ経由で上書きされる。
// 呼び出し例: go build -ldflags "-X main.version=$GITHUB_SHA".
// ここでは fmt 経由で空文字をシンプルなフォールバックにしておく。
func BuildInfo(version, env string) string {
	if version == "" {
		version = "dev"
	}
	if env == "" {
		env = "development"
	}
	return fmt.Sprintf("%s@%s", version, env)
}
