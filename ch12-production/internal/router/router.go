package router

import (
	"log/slog"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/forest6511/go-web-textbook-examples/ch12-production/internal/auth"
	"github.com/forest6511/go-web-textbook-examples/ch12-production/internal/handler"
	mw "github.com/forest6511/go-web-textbook-examples/ch12-production/internal/middleware"
)

type Deps struct {
	Logger            *slog.Logger
	RateLimiter       *mw.IPRateLimiter
	TaskHandler       *handler.TaskHandler
	AuthHandler       *handler.AuthHandler
	AttachmentHandler *handler.AttachmentHandler
	HealthHandler     *handler.HealthHandler
	AuthCfg           *auth.Config
	Production        bool
	// SentryEnabled が true のときだけ sentrygin ミドルウェアを有効化
	SentryEnabled bool
}

func New(d Deps) *gin.Engine {
	r := gin.New()

	r.Use(mw.Recovery(d.Logger))
	r.Use(mw.RequestID())
	if d.SentryEnabled {
		// Repanic=true: Gin の Recovery に panic を再 throw し、両方の経路を残す
		// WaitForDelivery=false: リクエストパスでは待たず、プロセス終了時に Flush でまとめて送る
		r.Use(sentrygin.New(sentrygin.Options{
			Repanic:         true,
			WaitForDelivery: false,
			Timeout:         2 * time.Second,
		}))
	}
	r.Use(mw.Logger(d.Logger))
	r.Use(mw.SecurityHeaders(d.Production))
	r.Use(cors.New(corsConfig()))
	r.Use(d.RateLimiter.Middleware())
	r.Use(gzipMiddleware())
	r.Use(mw.Errors(d.Logger)) // innermost: handler 返り直後に c.Errors を処理

	// healthz / readyz: LB の死活判定。認証・レート制限の影響を受けない
	if d.HealthHandler != nil {
		r.GET("/healthz", d.HealthHandler.Liveness)
		r.GET("/readyz", d.HealthHandler.Readiness)
	} else {
		// 後方互換: HealthHandler 未設定時は旧挙動
		r.GET("/healthz", func(c *gin.Context) { c.String(200, "ok") })
	}
	// /metrics: Prometheus スクレイプ用。default registry を使う
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := r.Group("/api/v1")

	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/signup", d.AuthHandler.Signup)
		authGroup.POST("/login", d.AuthHandler.Login)
		authGroup.POST("/refresh", d.AuthHandler.Refresh)
		authGroup.POST("/logout", d.AuthHandler.Logout)
	}

	protected := v1.Group("/", auth.JWTAuth(d.AuthCfg))
	{
		registerTaskRoutes(protected, d.TaskHandler)
		registerAttachmentRoutes(protected, d.AttachmentHandler)
	}
	return r
}

func registerAttachmentRoutes(
	g *gin.RouterGroup, h *handler.AttachmentHandler,
) {
	a := g.Group("/attachments")
	a.POST("", h.Upload)
	a.POST("/presign", h.PresignUpload)
	a.GET("/:id/download", h.GetDownloadURL)
}

func registerTaskRoutes(g *gin.RouterGroup, h *handler.TaskHandler) {
	tasks := g.Group("/tasks")
	tasks.POST("", h.Create)
	tasks.GET("", h.List)
	tasks.GET("/:id", h.Get)
	tasks.PATCH("/:id/status", h.UpdateStatus)
	tasks.DELETE("/:id", h.Delete)
}

func corsConfig() cors.Config {
	return cors.Config{
		AllowOrigins: []string{
			"https://app.example.com",
			"http://localhost:5173",
		},
		AllowMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
		},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}

func gzipMiddleware() gin.HandlerFunc {
	return gzip.Gzip(
		gzip.DefaultCompression,
		gzip.WithExcludedExtensions([]string{
			".png", ".jpg", ".jpeg", ".webp", ".pdf", ".mp4",
		}),
		gzip.WithExcludedPaths([]string{"/healthz", "/readyz", "/metrics"}),
		gzip.WithMinLength(1024),
	)
}
