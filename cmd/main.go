package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pressly/goose/v3"
	"github.com/zechao/faceit-user-svc/config"
	api "github.com/zechao/faceit-user-svc/http"
	"github.com/zechao/faceit-user-svc/postgres"
	"github.com/zechao/faceit-user-svc/service"
	"github.com/zechao/faceit-user-svc/tracing"
)

func main() {
	log.Println("App running in environment:", config.ENVs.APPEnv)
	db, err := postgres.NewPostgreStorage(postgres.Config{
		DBUser:     config.ENVs.DBUser,
		DBHost:     config.ENVs.DBHost,
		DBName:     config.ENVs.DBName,
		DBPassword: config.ENVs.DBPassword,
		DBPort:     config.ENVs.DBPort,
		DBSSLMode:  config.ENVs.DBSSLMode,
		DebugMode:  config.ENVs.DebugMode,
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

	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(tracing.TracingMiddleware())

	userStore := postgres.NewUserRepository(db)
	userService := service.NewUserService(userStore)
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
