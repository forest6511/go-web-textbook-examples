package router

import (
	"log/slog"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"github.com/forest6511/go-web-textbook-examples/ch08-file-upload/internal/auth"
	"github.com/forest6511/go-web-textbook-examples/ch08-file-upload/internal/handler"
	mw "github.com/forest6511/go-web-textbook-examples/ch08-file-upload/internal/middleware"
)

type Deps struct {
	Logger            *slog.Logger
	RateLimiter       *mw.IPRateLimiter
	TaskHandler       *handler.TaskHandler
	AuthHandler       *handler.AuthHandler
	AttachmentHandler *handler.AttachmentHandler
	AuthCfg           *auth.Config
	Production        bool
}

func New(d Deps) *gin.Engine {
	r := gin.New()

	// healthz はミドルウェアチェーンの前に登録する。CORS / レート制限 / Gzip
	// の影響を受けずに常に 200 を返せるようにし、Cloud Run や Kubernetes の
	// プローブが確実に疎通するようにする。
	r.GET("/healthz", func(c *gin.Context) {
		c.String(200, "ok")
	})

	r.Use(mw.Recovery(d.Logger))
	r.Use(mw.RequestID())
	r.Use(mw.Errors(d.Logger))
	r.Use(mw.Logger(d.Logger))
	r.Use(mw.SecurityHeaders(d.Production))
	r.Use(cors.New(corsConfig()))
	r.Use(d.RateLimiter.Middleware())
	r.Use(gzipMiddleware())

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
		gzip.WithExcludedPaths([]string{"/healthz", "/metrics"}),
		gzip.WithMinLength(1024),
	)
}
