package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pressly/goose"
	"github.com/zechao/faceit-user-svc/config"
	"github.com/zechao/faceit-user-svc/event"
	api "github.com/zechao/faceit-user-svc/http"
	"github.com/zechao/faceit-user-svc/log"
	"github.com/zechao/faceit-user-svc/postgres"
	"github.com/zechao/faceit-user-svc/service"
	"github.com/zechao/faceit-user-svc/tracing"

	"gorm.io/gorm"
)

func main() {
	log.Println("Starting application in " + config.ENVs.APPEnv + " mode")
	db, err := setupDatabase()
	if err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	natConn, err := event.NewNatConnection(fmt.Sprintf("nats://%s:%s", config.ENVs.NatsConfig.NatsHost, config.ENVs.NatsConfig.NatsPort))
	if err != nil {
		log.Fatalf("failed to connect to NATS: %v", err)
	}
	defer natConn.Close()

	eventHandler := event.NewNatsEventHandler(natConn, config.ENVs.NatsConfig.Topic)
	//Subscribe to the event bus to simulate another service
	eventHandler.Subscribe(func(event event.Event) {
		data, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal event: %v", err)
		}
		log.Printf("Received event: %v", string(data))
	})

	if config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	// recover from any panics and return a 500 error
	router.Use(gin.Recovery())
	// tracing middleware to set traceID in context
	router.Use(tracing.TracingMiddleware())
	// log request and response
	router.Use(log.GinLoggerMiddleware())

	// add health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	userStore := postgres.NewUserRepository(db)
	userService := service.NewUserService(userStore, eventHandler)
	userHandler := api.NewUserHandler(userService)
	userHandler.RegisterRoutes(router)

	log.Println("Listening on port:", config.ENVs.HTTPPort)
	srv := &http.Server{
		Addr:    ":" + config.ENVs.HTTPPort,
		Handler: router.Handler(),
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown: %v", err)
	}
	// catching ctx.Done(). timeout of 5 seconds.
	select {
	case <-ctx.Done():
		log.Println("timeout of 5 seconds.")
	}
	log.Println("Server exiting")
}

func setupDatabase() (*gorm.DB, error) {
	db, err := postgres.NewPostgreStorage(config.DBConfig{
		DBUser:     config.ENVs.DBConfig.DBUser,
		DBHost:     config.ENVs.DBConfig.DBHost,
		DBName:     config.ENVs.DBConfig.DBName,
		DBPassword: config.ENVs.DBConfig.DBPassword,
		DBPort:     config.ENVs.DBConfig.DBPort,
		DBSSLMode:  config.ENVs.DBConfig.DBSSLMode,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}

	conn, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying db connection: %w", err)
	}
	if err := goose.Up(conn, "./migrations"); err != nil {
		return nil, fmt.Errorf("failed to execute migrations: %w", err)
	}

	return db, nil
}
