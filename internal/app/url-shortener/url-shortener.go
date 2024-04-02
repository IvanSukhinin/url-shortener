package url_shortener

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"url-shortener/internal/clients/sso/ssogrpc"
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/url/aliaslist"
	"url-shortener/internal/http-server/handlers/url/del"
	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/http-server/middleware/auth"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage/postgresql"
)

const envDev = "dev"

type UrlShortener struct {
	cfg    *config.Config
	log    *slog.Logger
	db     *postgresql.Storage
	sso    *ssogrpc.Client
	router *chi.Mux
	srv    *http.Server
}

func New() *UrlShortener {
	us := &UrlShortener{}
	us.cfg = config.MustLoad()
	us.log = setupLogger(us.cfg.Env)
	us.db = initDatabase(us)
	api, err := initSsoGrpcApi(us)
	if err != nil {
		panic(err)
	}
	us.sso = api
	us.router = initRouter(us)
	us.srv = initServer(us)
	return us
}

func Run(us *UrlShortener) {
	defer us.db.Close()

	us.log.Info(
		"starting url-shortener",
		slog.String("env", us.cfg.Env),
		slog.Any("cfg", us.cfg),
	)
	us.log.Debug("debug messages are enabled")
	us.log.Info("starting server", slog.String("address", us.cfg.HTTPServer.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := us.srv.ListenAndServe(); err != nil {
			us.log.Error("failed to start server")
		}
	}()
	us.log.Info("server started")

	<-done
	us.log.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := us.srv.Shutdown(ctx); err != nil {
		us.log.Error("failed to stop server", sl.Err(err))
		return
	}

	us.log.Info("server stopped")
}

func initDatabase(us *UrlShortener) *postgresql.Storage {
	var err error
	db, err := postgresql.New(us.cfg.Db)
	if err != nil {
		us.log.Error("failed to init database", sl.Err(err))
		os.Exit(1)
	}
	return db
}

func initSsoGrpcApi(us *UrlShortener) (*ssogrpc.Client, error) {
	return ssogrpc.New(
		context.Background(),
		us.log,
		us.cfg.SsoGrpcApi.Address,
		us.cfg.SsoGrpcApi.Timeout,
		us.cfg.SsoGrpcApi.RetriesCount,
	)
}

func initRouter(us *UrlShortener) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(auth.New(us.log, us.cfg.App.Secret, us.sso))

	initHandlers(router, us)

	return router
}

func initHandlers(router *chi.Mux, us *UrlShortener) {
	router.Get("/", alist.New(us.log, us.db))
	router.Post("/save", save.New(us.cfg, us.log, us.db))
	router.Get("/del", del.New(us.log, us.db))
	router.Get("/{alias}", redirect.New(us.log, us.db))
}

func initServer(us *UrlShortener) *http.Server {
	return &http.Server{
		Addr:         us.cfg.HTTPServer.Address,
		Handler:      us.router,
		ReadTimeout:  us.cfg.HTTPServer.Timeout,
		WriteTimeout: us.cfg.HTTPServer.Timeout,
		IdleTimeout:  us.cfg.HTTPServer.IdleTimeout,
	}
}

func setupLogger(env string) *slog.Logger {
	level := slog.LevelInfo
	if env == envDev {
		level = slog.LevelDebug
	}
	return slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}),
	)
}
