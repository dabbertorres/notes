package main

import (
	"net"
	"net/http"
	"time"

	"github.com/dabbertorres/notes/config"
	"github.com/dabbertorres/notes/internal/log"
	notesapiv1 "github.com/dabbertorres/notes/internal/notes/apiv1"
	"github.com/dabbertorres/notes/internal/scope"
	"github.com/dabbertorres/notes/util"
	"github.com/felixge/httpsnoop"
	"github.com/google/uuid"
	"github.com/samber/do"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func setupServer(injector *do.Injector) (*http.Server, error) {
	cfg := do.MustInvoke[*config.Config](injector)
	logger := do.MustInvoke[*zap.Logger](injector)

	mux := http.NewServeMux()

	notesService := do.MustInvoke[notesapiv1.Service](injector)

	mux.HandleFunc("POST /api/v1/notes", notesapiv1.PostNote(notesService))
	mux.HandleFunc("PUT /api/v1/notes/{id}", notesapiv1.PutNote(notesService))
	mux.HandleFunc("DELETE /api/v1/notes/{id}", notesapiv1.DeleteNote(notesService))
	mux.HandleFunc("GET /api/v1/notes/{id}", notesapiv1.GetNote(notesService))
	mux.HandleFunc("GET /api/v1/notes", notesapiv1.ListNotes(notesService))

	mux.HandleFunc("POST /api/v1/tags", notesapiv1.PostTag(notesService))
	mux.HandleFunc("PUT /api/v1/tags", notesapiv1.PutTag(notesService))
	mux.HandleFunc("DELETE /api/v1/tags/{id}", notesapiv1.DeleteTag(notesService))
	mux.HandleFunc("GET /api/v1/tags/{id}", notesapiv1.GetTag(notesService))
	mux.HandleFunc("GET /api/v1/tags", notesapiv1.ListTags(notesService))

	mux.HandleFunc("POST /api/v1/users", nil)
	mux.HandleFunc("PUT /api/v1/users/{id}", nil)
	mux.HandleFunc("POST /api/v1/users/{id}/session", nil)
	mux.HandleFunc("DELETE /api/v1/users/{id}/session", nil)

	serverLog := util.Must(zap.NewStdLogAt(logger.Named("server"), zapcore.ErrorLevel))

	mw := chainMiddleware(
		tracingMiddleware(logger),
		loggingMiddleware(),
		recoveryMiddleware(),
	)

	srv := &http.Server{
		Addr:                         cfg.HTTP.Addr,
		Handler:                      mw(mux),
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    nil,
		ReadTimeout:                  0,
		ReadHeaderTimeout:            2 * time.Second,
		WriteTimeout:                 0,
		IdleTimeout:                  5 * time.Minute,
		MaxHeaderBytes:               int(cfg.HTTP.MaxHeaderBytes),
		TLSNextProto:                 nil,
		ConnState:                    nil,
		ErrorLog:                     serverLog,
		BaseContext:                  nil,
		ConnContext:                  nil,
	}

	if cfg.HTTP.LogConnections {
		srv.ConnState = func(c net.Conn, cs http.ConnState) {
		}
	}

	return srv, nil
}

type middleware func(next http.Handler) http.Handler

func chainMiddleware(mw ...middleware) middleware {
	if len(mw) == 0 {
		return func(next http.Handler) http.Handler { return next }
	}

	if len(mw) == 1 {
		return mw[0]
	}

	chained := mw[len(mw)-1]
	for i := len(mw) - 2; i >= 0; i-- {
		nextMW := chained
		chained = func(next http.Handler) http.Handler { return mw[i](nextMW(next)) }
	}

	return chained
}

func tracingMiddleware(logger *zap.Logger) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := uuid.New()
			ctx := scope.WithRequestID(r.Context(), id)
			ctx = scope.WithLogger(ctx, logger.With(zap.Stringer("request_id", id)))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func loggingMiddleware() middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			metrics := httpsnoop.CaptureMetrics(next, w, r)

			log.Info(r.Context(), "request/response",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int64("content_length", r.ContentLength),
				zap.String("user_agent", r.UserAgent()),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("referer", r.Referer()),
				zap.String("protocol", r.Proto),
				zap.Int("status", metrics.Code),
				zap.Int64("latency", metrics.Duration.Milliseconds()),
				zap.Int64("response_size", metrics.Written),
			)
		})
	}
}

func recoveryMiddleware() middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Error(r.Context(), "panic!", zap.Any("error", err))
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
