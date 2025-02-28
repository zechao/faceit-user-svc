package main

import (
	"embed"
	"log"
	"os"

	"github.com/pressly/goose/v3"
	"github.com/zechao/faceit-user-svc/config"
	"github.com/zechao/faceit-user-svc/postgres"
)

//go:embed *.sql
var embedMigrations embed.FS

// RunMigration runs the migration for the database for given direction
func main() {
	runMigration(config.DBConfig{
		DBUser:     config.ENVs.DBConfig.DBUser,
		DBHost:     config.ENVs.DBConfig.DBHost,
		DBName:     config.ENVs.DBConfig.DBName,
		DBPassword: config.ENVs.DBConfig.DBPassword,
		DBPort:     config.ENVs.DBConfig.DBPort,
		DBSSLMode:  config.ENVs.DBConfig.DBSSLMode,
	})
}

func runMigration(config config.DBConfig) {
	db, err := postgres.NewPostgreStorage(config)
	if err != nil {
		log.Fatalf("wrong db config: %v", err)
	}
	sqldb, err := db.DB()
	if err != nil {
		log.Fatalf("unable to connect to the database: %v", err)
	}

	log.Println("runnig migration")
	goose.SetBaseFS(embedMigrations)

	direction := ""
	if len(os.Args) > 1 {
		direction = os.Args[1]
	}
	switch direction {
	case "up":
		if err := goose.Up(sqldb, "."); err != nil {
			log.Panic(err)
		}
	case "down":
		if err := goose.Down(sqldb, "."); err != nil {
			log.Panic(err)
		}
	default:
		log.Fatal("Please, use up or down")
	}
}
