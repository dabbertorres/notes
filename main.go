package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/dabbertorres/notes/config"
	"github.com/dabbertorres/notes/internal/notes"
	"github.com/dabbertorres/notes/internal/notes/apiv1"
	"github.com/samber/do"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "cfg", "config.json", "Path to config file.")
	flag.Parse()

	injector := do.NewWithOpts(&do.InjectorOpts{})

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, unix.SIGTERM)
	defer cancel()

	do.ProvideValue(injector, ctx)
	do.ProvideNamedValue(injector, config.PathName, configPath)
	do.Provide(injector, config.Load)
	do.Provide(injector, setupLogging)
	do.Provide(injector, setupDatabase)
	do.Provide(injector, setupServer)
	do.Provide(injector, func(i *do.Injector) (apiv1.Service, error) { return notes.NewService(i) })

	logger := do.MustInvoke[*zap.Logger](injector)
	defer logger.Sync()

	srv := do.MustInvoke[*http.Server](injector)

	logger.Info("starting")

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("error running server", zap.Error(err))
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
	defer shutdownCancel()

	logger.Info("shutting down")

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("error shutting down server:", zap.Error(err))
	}
}
