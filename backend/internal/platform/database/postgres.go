package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"apihorpug/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(cfg config.Config) (*pgxpool.Pool, error) {
	if err := ensureDatabaseExists(cfg); err != nil {
		return nil, err
	}

	poolCfg, err := pgxpool.ParseConfig(dsn(cfg, cfg.DBName))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func dsn(cfg config.Config, dbName string) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s timezone=Asia/Bangkok",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUsername,
		cfg.DBPassword,
		dbName,
		cfg.DBSSLMode,
	)
}

// ensureDatabaseExists creates cfg.DBName via the maintenance "postgres"
// database if it doesn't exist yet. Postgres has no "CREATE DATABASE IF NOT
// EXISTS", so this checks pg_database first and only creates it when missing.
//
// It tries cfg.DBName directly first and returns early on success, since
// connection poolers such as pgbouncer are commonly configured to proxy only
// the application database and reject connections to "postgres" itself. The
// maintenance connection is only needed the first time the database doesn't
// exist yet, which happens against the database directly (e.g. local dev),
// never through a pooler that already targets an existing database.
func ensureDatabaseExists(cfg config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if targetConn, err := pgx.Connect(ctx, dsn(cfg, cfg.DBName)); err == nil {
		targetConn.Close(ctx)
		return nil
	}

	adminConn, err := pgx.Connect(ctx, dsn(cfg, "postgres"))
	if err != nil {
		return fmt.Errorf("connect to maintenance database: %w", err)
	}
	defer adminConn.Close(ctx)

	var exists bool
	err = adminConn.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", cfg.DBName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check database existence: %w", err)
	}
	if exists {
		return nil
	}

	log.Printf("database %q does not exist, creating it", cfg.DBName)

	// Database names can't be parameterized; pgx.Identifier quotes it safely.
	_, err = adminConn.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{cfg.DBName}.Sanitize())
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42P04" {
			// duplicate_database: created concurrently by another process, ignore.
			return nil
		}
		return fmt.Errorf("create database %q: %w", cfg.DBName, err)
	}

	return nil
}
