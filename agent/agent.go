package agent

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/brontomc/bronto/agent/api"
	"github.com/brontomc/bronto/agent/instance"
	"github.com/brontomc/bronto/agent/instance/state"
	"github.com/docker/docker/client"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	docker, err := client.NewClientWithOpts(client.WithHostFromEnv())
	if err != nil {
		return fmt.Errorf("error while connecting to the docker engine: %w", err)
	}

	stateStore, err := state.NewBoltStateStore(os.Getenv("test.db"))
	if err != nil {
		return fmt.Errorf("error while constructing state store: %w", err)
	}
	defer stateStore.Close()

	instances := instance.NewInstances(ctx, docker, stateStore)

	e := echo.New()
	e.Pre(middleware.AddTrailingSlash())
	handlers := api.Handlers{Instances: instances}

	iGroup := e.Group("/instance/:id", handlers.VerifyInstanceExistsMiddleware)
	iGroup.GET("/", handlers.HandleGetInstance)

	e.HTTPErrorHandler = api.ErrorHandler(e.DefaultHTTPErrorHandler)

	go func() {
		err = e.Start(":3000")
		if err != nil && err != http.ErrServerClosed {
			log.Fatal("error while starting api server: %w", err)
		}
	}()
	defer e.Shutdown(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh
	cancel()

	slog.Info("Shuttding down")

	return nil
}
