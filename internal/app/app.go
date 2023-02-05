package app

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"vpn-wg/internal/config"
	"vpn-wg/internal/router"
	"vpn-wg/internal/server"
	"vpn-wg/internal/service"
	"vpn-wg/internal/store/jsondb"
)

func Run() {
	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}

	db, err := jsondb.New("./db", cfg.Server, cfg.Global)
	if err != nil {
		panic(err)
	}

	if err := db.Init(); err != nil {
		panic(err)
	}

	services := service.NewServices(db)

	newRouter := router.NewRouter(services)

	srv := server.NewServer(cfg.HTTP, newRouter.Init())

	go func() {
		if err := srv.Run(); !errors.Is(err, http.ErrServerClosed) {
			logrus.Errorf("error occurred while running http server: %s\n", err.Error())
		}
	}()

	logrus.Info("Server started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	const timeout = 5 * time.Second

	ctx, shutdown := context.WithTimeout(context.Background(), timeout)
	defer shutdown()

	if err := srv.Stop(ctx); err != nil {
		logrus.Errorf("failed to stop server: %v", err)
	}

}
