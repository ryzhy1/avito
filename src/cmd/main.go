package main

import (
	"flag"
	"git.codenrock.com/avito/internal/app"
	"git.codenrock.com/avito/internal/config"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envDev   = "dev"
	envProd  = "prod"
	envLocal = "local"
)

func main() {
	cfg := config.MustLoad()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	env := fetchEnvironment()

	log.Info("Starting http", "env", env)

	application := app.New(log, cfg.ServerAddress, cfg.StorageConn)

	go application.HTTPServer.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	log.Info("Application stopped", slog.String("signal", sign.String()))

	application.HTTPServer.Stop()
}

func fetchEnvironment() string {
	var env string

	flag.StringVar(&env, "env", "", "environment to run server")
	flag.Parse()

	return env
}
