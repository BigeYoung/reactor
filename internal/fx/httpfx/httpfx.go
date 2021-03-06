package httpfx

import (
	"context"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/yagehu/reactor/config"
	"github.com/yagehu/reactor/internal/fx/middlewarefx/logging"
)

var Module = fx.Options(
	fx.Provide(New),
	fx.Invoke(startServer),
)

type Params struct {
	fx.In

	Config            config.Config
	LoggingMiddleware *logging.Middleware
}

type Result struct {
	fx.Out

	Router *mux.Router
	Server *http.Server
}

func New(p Params) (Result, error) {
	router := mux.NewRouter()
	server := http.Server{
		Addr:    net.JoinHostPort(p.Config.HTTP.Host, p.Config.HTTP.Port),
		Handler: p.LoggingMiddleware.Apply(router),
	}

	return Result{
		Router: router,
		Server: &server,
	}, nil
}

func startServer(
	lifecycle fx.Lifecycle,
	shutdowner fx.Shutdowner,
	logger *zap.Logger,
	server *http.Server,
) {
	lifecycle.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				if err := server.ListenAndServe(); err != nil {
					if err == http.ErrServerClosed {
						logger.Info("HTTP server closed.")
					} else {
						logger.Error("HTTP server error.", zap.Error(err))
					}
					_ = shutdowner.Shutdown()
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
}
