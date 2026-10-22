package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"tasklet_api/internal"
	"tasklet_api/internal/db"
)

func listenAddress(port string) string {
	if port == "" {
		return ":8080"
	}
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}

func run(ctx context.Context) error {
	conn, err := db.ConnectDatabase()
	if err != nil {
		return err
	}
	defer conn.Close()
	migrateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	err = db.Migrate(migrateCtx, conn)
	cancel()
	if err != nil {
		return err
	}
	server := &http.Server{
		Addr: listenAddress(os.Getenv("PORT")), Handler: internal.ConfigurarRotas(conn),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
		WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second,
	}
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return fmt.Errorf("iniciar HTTP: %w", err)
	}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	slog.Info("Servidor iniciado", "endereco", listener.Addr().String())
	select {
	case err = <-done:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err = server.Shutdown(shutdownCtx); err != nil {
			server.Close()
			return err
		}
		err = <-done
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := run(ctx); err != nil {
		slog.Error("Falha ao executar API", "error", err)
		os.Exit(1)
	}
}
