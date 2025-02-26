package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/zechao/faceit-user-svc/config"
	api "github.com/zechao/faceit-user-svc/http"
	"github.com/zechao/faceit-user-svc/postgres"
	"github.com/zechao/faceit-user-svc/service"
	"gorm.io/gorm"
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

	checkDBConnection(db)

	router := gin.Default()

	userStore := postgres.NewUserRepository(db)
	userService := service.NewUserService(userStore)
	userHandler := api.NewUserHandler(userService)

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	userHandler.RegisterRoutes(router)

	host := config.ENVs.HTTPHost + ":" + config.ENVs.HTTPPort
	log.Println("Listening on:", host)

	router.Run(host)
}

func checkDBConnection(db *gorm.DB) {
	conn, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	err = conn.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("DB connected")
}
