package server

import (
	"net/http"

	"app/docs"
	"app/internal/config"
	"app/internal/features/ai"
	"app/internal/features/auth"
	"app/internal/features/post"
	"app/internal/middleware"
	"app/internal/response"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// NewRouter merakit seluruh rute. aiSvc dapat di-inject agar mudah diuji.
func NewRouter(cfg *config.Config, db *gorm.DB, aiSvc *ai.Service) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), middleware.CORS(), middleware.RateLimiter(6000))

	authH := &auth.Handler{Svc: &auth.Service{Repo: &auth.Repository{DB: db}, Secret: cfg.AppKey}}
	postH := &post.Handler{Svc: &post.Service{Repo: &post.Repository{DB: db}}}
	aiH := &ai.Handler{Svc: aiSvc}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/api.json")))
	r.GET("/api.json", func(c *gin.Context) {
		scheme := "http"
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		c.Data(http.StatusOK, "application/json", []byte(docs.GetSwaggerJSON(c.Request.Host, scheme)))
	})

	api := r.Group("/api")
	pub := api.Group("/public")
	pub.GET("/categories", postH.Categories)
	pub.GET("/posts", postH.PublicList)
	pub.GET("/posts/:slug", postH.PublicDetail)

	api.POST("/auth/register", authH.Register)
	api.POST("/auth/login", authH.Login)

	guard := api.Group("", middleware.Auth(cfg.AppKey))
	guard.POST("/auth/logout", authH.Logout)
	guard.GET("/auth/profile", authH.Profile)
	guard.PUT("/auth/profile", authH.UpdateProfile)

	guard.GET("/posts", postH.List)
	guard.POST("/posts", postH.Create)
	guard.GET("/posts/stats", postH.Stats)
	guard.POST("/posts/ai-generate", aiH.Generate)
	guard.GET("/posts/:id", postH.Detail)
	guard.PUT("/posts/:id", postH.Update)
	guard.PATCH("/posts/:id/toggle-publish", postH.Toggle)
	guard.DELETE("/posts/:id", postH.Delete)

	admin := guard.Group("/admin", middleware.RequireRole("admin"))
	admin.GET("/users", postH.AdminUsers)
	admin.GET("/posts", postH.AdminPosts)
	admin.GET("/users/:id/posts", postH.AdminUserPosts)

	r.NoRoute(func(c *gin.Context) {
		response.Fail(c, http.StatusNotFound, "Endpoint not found")
	})
	return r
}
