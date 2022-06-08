package server

import (
	"autograph-backend-controller/server/common"
	"autograph-backend-controller/server/handler"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Config struct {
	Host      string
	Port      int
	DebugMode bool
}

type Server struct {
	engine *gin.Engine
	config *Config
}

func New(config *Config) *Server {
	eng := gin.Default()

	eng.Use(common.LogRequest)
	eng.Use(common.SetUserInfo(config.DebugMode))
	eng.Use(cors.Default())

	eng.GET("/test/coffee", coffeeHandler)

	eng.GET("/fileinfo", handler.GetFileInfo)

	// 需要登录的路由
	adminGroup := eng.Group("admin")
	{
		adminGroup.Use(common.RejectNotLogin(config.DebugMode))

		adminGroup.POST("/upload", handler.UploadFile)
		adminGroup.POST("/build", handler.BuildVersion)
		adminGroup.GET("/listfile", handler.ListFile)
		adminGroup.GET("/listversion", handler.ListVersion)
		adminGroup.GET("/listextractor", handler.ListExtractor)
		adminGroup.POST("/intervention", handler.Intervention)
		adminGroup.GET("/search", handler.Search)
	}

	return &Server{
		engine: eng,
		config: config,
	}
}

func (s *Server) RunServer() error {
	return s.engine.Run(fmt.Sprintf("%s:%d", s.config.Host, s.config.Port))
}
