package main

import (
	"net"
	"net/http"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/dabbertorres/notes/internal/common/apiv1"
	"github.com/dabbertorres/notes/internal/config"
	"github.com/dabbertorres/notes/internal/log"
	notesapiv1 "github.com/dabbertorres/notes/internal/notes/apiv1"
	"github.com/dabbertorres/notes/internal/scope"
	"github.com/dabbertorres/notes/internal/telemetry"
	usersapiv1 "github.com/dabbertorres/notes/internal/users/apiv1"
	"github.com/dabbertorres/notes/internal/util"
)

func setupServer(injector do.Injector) (*http.Server, error) {
	do.MustInvoke[telemetry.Service](injector)

	mux := http.NewServeMux()

	mux.Handle("GET /healthz", healthCheck(injector))

	logLevel := do.MustInvoke[zap.AtomicLevel](injector)
	mux.Handle("PUT /logging", logLevel)
	mux.Handle("GET /logging", logLevel)

	notesService := do.MustInvokeAs[notesapiv1.Service](injector)

	addHandler(mux, "POST", "/api/v1/notes", notesapiv1.PostNote(notesService))
	addHandler(mux, "PUT", "/api/v1/notes/{id}", notesapiv1.PutNote(notesService))
	addHandler(mux, "DELETE", "/api/v1/notes/{id}", notesapiv1.DeleteNote(notesService))
	addHandler(mux, "GET", "/api/v1/notes/{id}", notesapiv1.GetNote(notesService))
	addHandler(mux, "GET", "/api/v1/notes", notesapiv1.ListNotes(notesService))

	addHandler(mux, "POST", "/api/v1/tags", notesapiv1.PostTag(notesService))
	addHandler(mux, "PUT", "/api/v1/tags", notesapiv1.PutTag(notesService))
	addHandler(mux, "DELETE", "/api/v1/tags/{id}", notesapiv1.DeleteTag(notesService))
	addHandler(mux, "GET", "/api/v1/tags/{id}", notesapiv1.GetTag(notesService))
	addHandler(mux, "GET", "/api/v1/tags", notesapiv1.ListTags(notesService))

	usersService := do.MustInvokeAs[usersapiv1.Service](injector)

	addHandler(mux, "POST", "/api/v1/users", usersapiv1.PostUser(usersService))
	addHandler(mux, "PUT", "/api/v1/users/{id}", usersapiv1.PutUser(usersService))
	addHandler(mux, "POST", "/api/v1/users/{id}/session", usersapiv1.PostSession(usersService))
	addHandler(mux, "DELETE", "/api/v1/users/{id}/session", usersapiv1.DeleteSession(usersService))

	mw := util.ChainReverse1(
		otelhttp.NewMiddleware("server",
			otelhttp.WithFilter(func(r *http.Request) bool {
				switch r.URL.Path {
				case "/healthz":
					return false
				case "/logging":
					return false
				default:
					return true
				}
			}),
			otelhttp.WithMessageEvents(
				otelhttp.ReadEvents,
				otelhttp.WriteEvents,
			),
			otelhttp.WithPublicEndpoint(),
			otelhttp.WithServerName("notes"),
		),
		loggingMiddleware(),
		recoveryMiddleware(),
		// TODO: auth middleware
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := scope.WithUserID(r.Context(), uuid.Max)

				next.ServeHTTP(w, r.WithContext(ctx))
			})
		},
	)

	cfg := do.MustInvoke[*config.Config](injector)
	logger := do.MustInvoke[*zap.Logger](injector)

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
		ErrorLog:                     util.Must(zap.NewStdLogAt(logger.Named("server"), zapcore.ErrorLevel)),
		BaseContext:                  nil,
		ConnContext:                  nil,
	}

	if cfg.HTTP.LogConnections {
		srv.ConnState = func(c net.Conn, cs http.ConnState) {
		}
	}

	return srv, nil
}

func healthCheck(injector do.Injector) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		results := injector.HealthCheckWithContext(r.Context())

		var hasErrors bool
		for _, e := range results {
			if e != nil {
				hasErrors = true
				break
			}
		}

		status := util.FoldBool(hasErrors, http.StatusInternalServerError, http.StatusOK)
		apiv1.WriteJSON(r.Context(), w, status, results)
	})
}

// addHandler is a helper function for adding a [http.Handler] to mux, tagging it with OTEL's http.route attribute,
// and setting the span's name correctly.
func addHandler(mux *http.ServeMux, method, route string, handler http.Handler) {
	spanName := method + " " + route

	next := otelhttp.WithRouteTag(route, handler)

	mux.HandleFunc(spanName, func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())
		span.SetName(spanName)

		next.ServeHTTP(w, r)
	})
}

type middleware = func(http.Handler) http.Handler

func loggingMiddleware() middleware {
	logger := zap.L()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := trace.SpanContextFromContext(r.Context())
			if span.IsValid() {
				requestLog := logger.With(
					zap.String("trace", span.TraceID().String()),
					zap.String("span", span.SpanID().String()),
				)

				ctx := scope.WithLogger(r.Context(), requestLog)
				r = r.WithContext(ctx)
			}

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
