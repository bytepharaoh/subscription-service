package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/bytepharaoh/subscription-service/internal/handler"
	"github.com/bytepharaoh/subscription-service/internal/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	readTimeout     = 10 * time.Second
	writeTimeout    = 10 * time.Second
	idleTimeout     = 60 * time.Second
	shutdownTimeout = 5 * time.Second
	requestTimeout  = 8 * time.Second
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

func New(port string, h *handler.Handler, logger *slog.Logger, appEnv string) *Server {
	if appEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// global middleware
	router.Use(
		middleware.Recovery(logger),
		middleware.RequestID(),
		middleware.Logger(logger),
		middleware.CORS(),
		middleware.Timeout(requestTimeout),
	)

	// routes
	registerRoutes(router, h)

	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%s", port),
			Handler:      router,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
		logger: logger,
	}
}

func registerRoutes(router *gin.Engine, h *handler.Handler) {
	router.GET("/health", h.Health)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	{
		subs := v1.Group("/subscriptions")
		{
			subs.POST("", h.Create)
			subs.GET("", h.List)
			subs.GET("/total-cost", h.GetTotalCost)
			subs.GET("/:id", h.GetByID)
			subs.PUT("/:id", h.Update)
			subs.DELETE("/:id", h.Delete)
		}
	}
}

func (s *Server) Run() error {
	s.logger.Info("server starting", slog.String("addr", s.httpServer.Addr))
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("server shutting down gracefully")
	return s.httpServer.Shutdown(ctx)
}
