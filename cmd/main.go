package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pressly/goose/v3"
	"github.com/zechao/faceit-user-svc/config"
	"github.com/zechao/faceit-user-svc/event"
	api "github.com/zechao/faceit-user-svc/http"
	"github.com/zechao/faceit-user-svc/postgres"
	"github.com/zechao/faceit-user-svc/service"
	"github.com/zechao/faceit-user-svc/tracing"
)

func main() {
	log.Println("App running in environment:", config.ENVs.APPEnv)
	db, err := postgres.NewPostgreStorage(config.DBConfig{
		DBUser:     config.ENVs.DBConfig.DBUser,
		DBHost:     config.ENVs.DBConfig.DBHost,
		DBName:     config.ENVs.DBConfig.DBName,
		DBPassword: config.ENVs.DBConfig.DBPassword,
		DBPort:     config.ENVs.DBConfig.DBPort,
		DBSSLMode:  config.ENVs.DBConfig.DBSSLMode,
		DebugMode:  config.ENVs.DBConfig.DebugMode,
	})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	conn, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlaying db connection: %v", err)
	}
	if err := goose.Up(conn, "./migrations"); err != nil {
		log.Fatalf("Failed to execute migrations: %v", err)
	}

	natConn, err := event.NewNatConnection(fmt.Sprintf("nats://%s:%s", config.ENVs.NatsConfig.NatsHost, config.ENVs.NatsConfig.NatsPort))
	if err != nil {
		log.Fatalf("failed to connect to NATS: %v", err)
	}
	defer natConn.Close()

	eventHandler := event.NewNatsEventHandler(natConn, config.ENVs.NatsConfig.Topic)

	// Subscribe to the event bus to simulate another service
	eventHandler.Subscribe(func(event event.Event) {
		data, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal event: %v", err)
		}
		log.Printf("Received event: %v", string(data))
	})

	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(tracing.TracingMiddleware())

	userStore := postgres.NewUserRepository(db)
	userService := service.NewUserService(userStore, eventHandler)
	userHandler := api.NewUserHandler(userService)

	router.Use(gin.Logger())
	// recover from any panics and return a 500 error
	router.Use(gin.Recovery())

	userHandler.RegisterRoutes(router)

	host := config.ENVs.HTTPHost + ":" + config.ENVs.HTTPPort
	log.Println("Listening on:", host)

	if err := router.Run(host); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to run server: %v", err)
	}

}
