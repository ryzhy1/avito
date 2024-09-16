package app

import (
	"git.codenrock.com/avito/internal/app/http-server"
	"git.codenrock.com/avito/internal/handlers"
	"git.codenrock.com/avito/internal/repository/postgres"
	"git.codenrock.com/avito/internal/routes"
	"git.codenrock.com/avito/internal/services"
	"github.com/gin-gonic/gin"
	"log/slog"
)

type App struct {
	HTTPServer *http_server.Server
}

func New(log *slog.Logger, serverPort, storagePath string) *App {
	storage, err := postgres.New(storagePath)
	if err != nil {
		panic(err)
	}

	tenderService := services.NewTenderService(log, storage)
	tenderHandler := handlers.NewTenderHandler(log, tenderService)

	bidService := services.NewBidService(log, storage)
	bidHandler := handlers.NewBidHandler(log, bidService)

	r := gin.Default()
	err = r.SetTrustedProxies(nil)
	if err != nil {
		return nil
	}
	routes.InitRoutes(r, tenderHandler, bidHandler)

	server := http_server.NewServer(log, serverPort, r)

	return &App{
		HTTPServer: server,
	}
}
