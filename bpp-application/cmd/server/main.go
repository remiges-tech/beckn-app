package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/remiges-tech/alya/service"
	"github.com/remiges-tech/logharbour/logharbour"

	dbsqlc "github.com/ion/winroom/bpp/db/sqlc"
	"github.com/ion/winroom/bpp/internal/config"
	"github.com/ion/winroom/bpp/internal/server"
)

func main() {
	// Load .env in non-production environments.
	// In production, variables are injected by the orchestrator (k8s, etc.).
	if err := godotenv.Load(); err != nil {
		// Not fatal — env vars may already be set in the environment.
		fmt.Fprintln(os.Stderr, "no .env file found, using process environment")
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	// --- Logging ---------------------------------------------------------
	// logharbour emits structured JSON logs to stdout.
	// The LoggerContext is shared across clones, allowing runtime priority
	// changes without recreating loggers.
	lctx := logharbour.NewLoggerContext(logLevel(cfg.LogLevel))
	lh := logharbour.NewLogger(lctx, "bpp", os.Stdout)

	lh.WithModule("main").Log("BPP starting")

	// --- Database --------------------------------------------------------
	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL())
	if err != nil {
		lh.WithModule("main").Crit().Error(err).Log("failed to parse database URL")
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		lh.WithModule("main").Crit().Error(err).Log("failed to create database pool")
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		lh.WithModule("main").Crit().Error(err).Log("database ping failed")
		os.Exit(1)
	}
	lh.WithModule("main").Log("database connected")

	// --- BPP identity seed -----------------------------------------------
	// Every table with a bpp_id column has a FK into bpp_identities.
	// Ensure this BPP's row exists before accepting any inbound requests.
	if err := registerBPPIdentity(ctx, pool, cfg); err != nil {
		lh.WithModule("main").Crit().Error(err).Log("failed to register BPP identity")
		os.Exit(1)
	}
	lh.WithModule("main").Log("BPP identity registered: " + cfg.BppID)

	// --- HTTP server -----------------------------------------------------
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	svc := service.NewService(r).
		WithLogHarbour(lh).
		WithDatabase(pool)

	server.RegisterRoutes(svc, cfg)

	// --- Graceful shutdown -----------------------------------------------
	addr := fmt.Sprintf(":%s", cfg.AppPort)
	lh.WithModule("main").Log("server listening on " + addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := r.Run(addr); err != nil {
			lh.WithModule("main").Crit().Error(err).Log("server stopped")
			os.Exit(1)
		}
	}()

	<-quit
	lh.WithModule("main").Log("shutdown signal received, exiting")
}

// registerBPPIdentity upserts this BPP's subscriber record into bpp_identities.
// All other tables (providers, catalogs, resources, offers …) carry a bpp_id FK
// that references this table, so it must exist before any writes are attempted.
func registerBPPIdentity(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config) error {
	q := dbsqlc.New(pool)
	return q.UpsertBPPIdentity(ctx, dbsqlc.UpsertBPPIdentityParams{
		BppID:      cfg.BppID,
		BppUri:     cfg.BppURI,
		NetworkIds: []string{cfg.NetworkID},
	})
}

// logLevel maps the LOG_LEVEL env string to a logharbour priority constant.
func logLevel(level string) logharbour.LogPriority {
	switch level {
	case "debug2":
		return logharbour.Debug2
	case "debug1":
		return logharbour.Debug1
	case "debug0":
		return logharbour.Debug0
	case "warn":
		return logharbour.Warn
	case "err":
		return logharbour.Err
	case "crit":
		return logharbour.Crit
	case "sec":
		return logharbour.Sec
	default:
		return logharbour.Info
	}
}
