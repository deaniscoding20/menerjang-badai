package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/menerjangbadai/dashboard/internal/adapters/httpapi"
	oidcadapter "github.com/menerjangbadai/dashboard/internal/adapters/oidc"
	"github.com/menerjangbadai/dashboard/internal/adapters/postgres"
	"github.com/menerjangbadai/dashboard/internal/application"
	"github.com/menerjangbadai/dashboard/internal/platform"
)

// Set by -ldflags '-X main.buildSHA=<exact commit SHA>'; config may override it.
var buildSHA string

const usage = `Usage: mb-dashboard [serve|check-config]

serve         Validate configuration, ping the database, discover OIDC, then serve.
check-config  Validate configuration without network access or starting a listener.

With no command, serve is selected. Configuration is mandatory in every environment.
Migrations are never automatic; use mb-admin migrate explicitly.
See docs/14-local-development.md. Use --help for this message.
`

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, out, stderr io.Writer) int {
	command, err := parseCommand(args)
	if err != nil {
		fmt.Fprintln(stderr, "mb-dashboard:", err)
		return 2
	}
	if command == "help" {
		fmt.Fprint(out, usage)
		return 0
	}
	c, err := platform.LoadRuntimeConfig(buildSHA)
	if err != nil {
		fmt.Fprintln(stderr, "mb-dashboard:", err)
		return 1
	}
	if command == "check-config" {
		fmt.Fprintln(out, "configuration valid (dependencies not checked)")
		return 0
	}
	if err := serve(ctx, c, stderr); err != nil {
		// All errors crossing this boundary are fixed messages, never underlying errors.
		fmt.Fprintln(stderr, "mb-dashboard:", err)
		return 1
	}
	return 0
}

func serve(ctx context.Context, c platform.Config, stderr io.Writer) error {
	startup, cancel := context.WithTimeout(ctx, 20*time.Second)
	pool, err := pgxpool.New(startup, c.DatabaseURL)
	if err != nil {
		cancel()
		return errors.New("database configuration rejected")
	}
	defer pool.Close()
	if err = pool.Ping(startup); err != nil {
		cancel()
		return errors.New("database unavailable; check local setup and explicit migrations")
	}
	idp, err := oidcadapter.New(startup, c.Issuer, c.ClientID, c.ClientSecret, c.PublicURL+"/auth/callback")
	cancel()
	if err != nil {
		return errors.New("OIDC discovery unavailable or issuer mismatch")
	}
	if ctx.Err() != nil {
		return nil
	}
	store := &postgres.Store{Pool: pool}
	handler := httpapi.New(httpapi.Config{PublicURL: c.PublicURL, CommitSHA: c.CommitSHA, WebDir: c.WebDir}, application.Brands{Repo: store}, store, idp, pool.Ping, httpapi.NewTaskHandler(application.Tasks{Repo: store}))
	server := &http.Server{
		Addr: c.Addr, Handler: handler,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 20 * time.Second,
		WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second,
		MaxHeaderBytes: 32 << 10,
		ErrorLog:       log.New(safeHTTPLog{stderr}, "", 0),
	}
	listener, err := net.Listen("tcp", c.Addr)
	if err != nil {
		return errors.New("HTTP listener unavailable")
	}
	defer listener.Close()
	fmt.Fprintln(stderr, "mb-dashboard: serving on configured loopback address")
	result := make(chan error, 1)
	go func() { result <- server.Serve(listener) }()
	select {
	case err := <-result:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return errors.New("HTTP server stopped unexpectedly")
	case <-ctx.Done():
		shutdown, done := context.WithTimeout(context.Background(), 10*time.Second)
		defer done()
		if err := server.Shutdown(shutdown); err != nil {
			_ = server.Close()
			<-result
			return errors.New("HTTP graceful shutdown deadline exceeded")
		}
		<-result
		fmt.Fprintln(stderr, "mb-dashboard: shutdown complete")
		return nil
	}
}

// The net/http error logger can include user-controlled input; do not forward it.
type safeHTTPLog struct{ out io.Writer }

func (w safeHTTPLog) Write(p []byte) (int, error) {
	_, err := fmt.Fprintln(w.out, "mb-dashboard: HTTP server diagnostic suppressed")
	return len(p), err
}

func parseCommand(args []string) (string, error) {
	if len(args) == 0 {
		return "serve", nil
	}
	if len(args) == 1 && (args[0] == "help" || args[0] == "--help" || args[0] == "-h") {
		return "help", nil
	}
	if args[0] == "serve" || args[0] == "check-config" {
		if len(args) == 1 {
			return args[0], nil
		}
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			return "help", nil
		}
	}
	return "", errors.New("invalid command arguments; use --help")
}
