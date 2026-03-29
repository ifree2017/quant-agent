package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	connString := getEnv("DATABASE_URL", "postgres://postgres:postgres@47.99.163.232:5432/quant_agent?sslmode=disable")
	migrationsPath := getEnv("MIGRATIONS_PATH", "file://migrations")

	m, err := migrate.New(migrationsPath, connString)
	if err != nil {
		log.Fatalf("failed to create migrate instance: %v", err)
	}

	// Graceful shutdown on SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		if err := m.Force(-1); err != nil {
			log.Printf("force migration rollback failed: %v", err)
		}
		os.Exit(1)
	}()

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migration failed: %v", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		log.Fatalf("failed to get version: %v", err)
	}

	if dirty {
		log.Fatalf("database is in a dirty state at version %d", version)
	}

	fmt.Printf("Migration complete. Current version: %d\n", version)
	fmt.Println("All tables and indices created successfully.")
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
