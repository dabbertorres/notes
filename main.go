package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/samber/do/v2"
	"go.uber.org/zap"

	"github.com/dabbertorres/notes/config"
	"github.com/dabbertorres/notes/internal/notes"
	"github.com/dabbertorres/notes/internal/telemetry"
	"github.com/dabbertorres/notes/internal/users"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "cfg", "config.json", "Path to config file.")
	flag.Parse()

	injector := do.NewWithOpts(&do.InjectorOpts{},
		notes.Package,
		users.Package,
		telemetry.Package,
	)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	do.ProvideValue(injector, ctx)
	do.ProvideValue(injector, zap.NewAtomicLevel())
	do.ProvideNamedValue(injector, config.PathName, configPath)
	do.Provide(injector, config.Load)
	do.Provide(injector, setupLogging)
	do.Provide(injector, setupDatabase)
	do.Provide(injector, setupServer)

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
