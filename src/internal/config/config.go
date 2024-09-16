package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	ServerAddress string
	StorageConn   string
}

func MustLoad() *Config {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		log.Fatal("SERVER_ADDRESS is not set")
	}

	postgresURL := os.Getenv("POSTGRES_CONN")
	if postgresURL == "" {
		log.Fatal("POSTGRES_CONN is not set")
	}

	return &Config{
		ServerAddress: serverAddress,
		StorageConn:   postgresURL,
	}
}
