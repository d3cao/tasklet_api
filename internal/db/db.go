package db

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func databaseURL() string {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	mode := os.Getenv("DB_SSLMODE")
	if mode == "" {
		mode = "disable"
	}
	u := url.URL{Scheme: "postgres", Host: net.JoinHostPort(host, port), User: url.UserPassword(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD")), Path: "/" + os.Getenv("DB_NAME")}
	q := url.Values{"sslmode": {mode}}
	u.RawQuery = q.Encode()
	return u.String()
}

func ConnectDatabase() (*sql.DB, error) {
	conn, err := sql.Open("pgx", databaseURL())
	if err != nil {
		return nil, fmt.Errorf("abrir banco: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = conn.PingContext(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("conectar banco: %w", err)
	}
	return conn, nil
}
