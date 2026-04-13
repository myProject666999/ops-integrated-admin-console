package runtime

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func setupRouter(srv *server) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddlewareGin())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	auth := r.Group("/api/auth")
	{
		auth.POST("/login", srv.handleLoginGin)
		auth.POST("/register", srv.handleRegisterGin)
		auth.POST("/logout", srv.authMiddleware(), srv.handleLogoutGin)
		auth.GET("/me", srv.authMiddleware(), srv.handleMeGin)
		auth.POST("/change-password", srv.authMiddleware(), srv.handleChangePasswordGin)
		auth.POST("/window-close-start", srv.authMiddleware(), srv.handleWindowCloseStartGin)
		auth.POST("/window-close-cancel", srv.authMiddleware(), srv.handleWindowCloseCancelGin)
	}

	projects := r.Group("/api/projects")
	projects.Use(srv.authMiddleware())
	{
		projects.GET("/credentials", srv.handleProjectCredentialsGin)
		projects.PUT("/credentials/:type", srv.handleProjectCredentialByTypeGin)
		projects.POST("/relogin", srv.handleProjectsReloginGin)
		projects.POST("/operate-async", srv.handleProjectOperateAsyncStartGin)
		projects.GET("/operate-async/:jobId", srv.handleProjectOperateAsyncStatusGin)
		projects.POST("/:type/load", srv.handleProjectLoadGin)
		projects.GET("/:type/batch-template", srv.handleProjectBatchTemplateGin)
		projects.POST("/:type/batch-upload", srv.handleProjectBatchUploadGin)
		projects.GET("/:type/batch-files", srv.handleProjectBatchFilesGin)
		projects.POST("/:type/operate", srv.handleProjectOperateGin)
	}

	r.GET("/api/logs", srv.authMiddleware(), srv.handleLogsGin)

	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "接口不存在"})
			return
		}
		srv.serveStaticGin(c)
	})

	return r
}

func corsMiddlewareGin() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func (s *server) serveStaticGin(c *gin.Context) {
	path := c.Request.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	filePath := filepath.Join(s.cfg.StaticDir, path)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		filePath = filepath.Join(s.cfg.StaticDir, "index.html")
	}

	c.File(filePath)
}
