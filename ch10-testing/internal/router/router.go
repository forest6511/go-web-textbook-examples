package router

import (
	"log/slog"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/forest6511/go-web-textbook-examples/ch10-testing/internal/auth"
	"github.com/forest6511/go-web-textbook-examples/ch10-testing/internal/handler"
	mw "github.com/forest6511/go-web-textbook-examples/ch10-testing/internal/middleware"
	"github.com/forest6511/go-web-textbook-examples/ch10-testing/internal/repository"
	"github.com/forest6511/go-web-textbook-examples/ch10-testing/internal/usecase"
)

type Deps struct {
	Logger            *slog.Logger
	RateLimiter       *mw.IPRateLimiter
	TaskHandler       *handler.TaskHandler
	AuthHandler       *handler.AuthHandler
	AttachmentHandler *handler.AttachmentHandler
	AuthCfg           *auth.Config
	Production        bool

	// テスト支援フィールド（Ch 10）。
	// Pool があり TaskHandler が未指定なら TaskHandler を自動構築する。
	// TaskRepo は Pool より優先され、pgxmock/fake を直接差し込む用途に使う。
	// AuthSkip=true では JWTAuth を回避し、固定 Principal(UserID=1) を注入する。
	Pool     *pgxpool.Pool
	TaskRepo usecase.TaskRepository
	AuthSkip bool
}

func New(d Deps) *gin.Engine {
	r := gin.New()

	logger := d.Logger
	if logger == nil {
		logger = slog.Default()
	}

	r.Use(mw.Recovery(logger))
	r.Use(mw.RequestID())
	r.Use(mw.Logger(logger))
	r.Use(mw.SecurityHeaders(d.Production))
	r.Use(cors.New(corsConfig()))
	if d.RateLimiter != nil {
		r.Use(d.RateLimiter.Middleware())
	}
	r.Use(gzipMiddleware())
	r.Use(mw.Errors(logger)) // innermost: handler 返り直後に c.Errors を処理

	r.GET("/healthz", func(c *gin.Context) {
		c.String(200, "ok")
	})
	v1 := r.Group("/api/v1")

	if d.AuthHandler != nil {
		authGroup := v1.Group("/auth")
		authGroup.POST("/signup", d.AuthHandler.Signup)
		authGroup.POST("/login", d.AuthHandler.Login)
		authGroup.POST("/refresh", d.AuthHandler.Refresh)
		authGroup.POST("/logout", d.AuthHandler.Logout)
	}

	var authMW gin.HandlerFunc
	if d.AuthSkip {
		authMW = testPrincipalMiddleware()
	} else {
		authMW = auth.JWTAuth(d.AuthCfg)
	}
	protected := v1.Group("/", authMW)
	{
		taskHandler := d.TaskHandler
		if taskHandler == nil && (d.TaskRepo != nil || d.Pool != nil) {
			taskHandler = buildTaskHandler(d)
		}
		if taskHandler != nil {
			registerTaskRoutes(protected, taskHandler)
		}
		if d.AttachmentHandler != nil {
			registerAttachmentRoutes(protected, d.AttachmentHandler)
		}
	}
	return r
}

// testPrincipalMiddleware は AuthSkip=true のとき固定 Principal を注入する。
// 本番コードでは使わない。ユニットテストで JWT 発行を省略するための補助。
func testPrincipalMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := auth.WithPrincipal(c.Request.Context(),
			auth.Principal{UserID: 1, Role: "user"})
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// buildTaskHandler は TaskRepo が優先、無ければ Pool から Repository を組む。
// TxRunner は Pool があれば本物、無ければ nil（Create/Get 系は repo 直呼びで動く）。
func buildTaskHandler(d Deps) *handler.TaskHandler {
	var repo usecase.TaskRepository
	if d.TaskRepo != nil {
		repo = d.TaskRepo
	} else {
		repo = repository.NewPostgresTaskRepo(d.Pool)
	}
	var tx *repository.TxRunner
	if d.Pool != nil {
		tx = repository.NewTxRunner(d.Pool)
	}
	uc := usecase.New(repo, tx)
	return handler.NewTaskHandler(uc)
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
