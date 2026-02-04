package main

import (
	"prac"
	config "prac/configs"
	"prac/pkg/handler"
	"prac/pkg/repository"
	"prac/pkg/service"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// @title       User Service API
// @version     1.0
// @description This API allows creating, retrieving, updating, and deleting users/products.

// @host     localhost:8080
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))
	if err := initConfig(); err != nil {
		logrus.Fatalf("error initializating configs: %s", err.Error())
	}

	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("error loading env variables: %s", err.Error())
	}

	redisConfig := config.RedisConfigFromViper()
	redisClient := config.NewRedisClient(redisConfig)

	dbConfig := repository.DBConfigFromViper()
	bd, err := repository.NewPostgresDB(dbConfig)

	if err != nil {
		logrus.Fatalf("failed to initialize db: %s", err.Error())
	}

	repository := repository.NewRepository(bd, redisClient)
	services := service.NewService(repository)
	handlers := handler.NewHandler(*services)
	srv := new(prac.Server)
	if err := srv.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil {
		logrus.Fatalf("error occured while running http server : %s", err.Error())
	}

}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.db", 0)
	return viper.ReadInConfig()
}
