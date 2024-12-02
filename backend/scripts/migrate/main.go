package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"backend/db"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// Define flags for command-line arguments
	operation := flag.String("op", "", "Operation to perform: up, down, create")
	migrationName := flag.String("name", "", "Name of the migration to create (required for create)")
	flag.Parse()

	_, filename, _, _ := runtime.Caller(0) // Gets this files path
	migrationsDir := filepath.Join(filepath.Dir(filename), "../../db/migrations")

	switch *operation {
	case "up":
		runMigrations(migrationsDir, "up")
	case "down":
		runMigrations(migrationsDir, "down")
	case "create":
		if *migrationName == "" {
			log.Fatalf("You must provide a migration name with -name for the create operation")
		}
		createMigration(migrationsDir, *migrationName)
	default:
		log.Fatalf("Usage: go run main.go -op <up|down|create> -name <Name of new migration (Only needed for create)>")
	}
}

func runMigrations(migrationsDir, direction string) {
	// Use your existing db package to get a database connection
	db.Establish_db_conn()
	conn := db.GetDb() // Adjust to match your package's method
	defer conn.Close()

	// Get the PostgreSQL driver instance
	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to create migration driver: %v", err)
	}

	// Create a new migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsDir),
		"postgres", driver,
	)
	if err != nil {
		log.Fatalf("Failed to initialize migrate instance: %v", err)
	}

	// Run migrations based on the direction
	if direction == "up" {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration up failed: %v", err)
		}
		log.Println("Migrations applied successfully")
	} else if direction == "down" {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration down failed: %v", err)
		}
		log.Println("Migrations rolled back successfully")
	}
}

func createMigration(migrationsDir, name string) {
	// Generate timestamp-based file names for the migration
	timestamp := time.Now().Format("20060102150405") // e.g., 20231201123456
	upFile := fmt.Sprintf("%s/%s_%s.up.sql", migrationsDir, timestamp, name)
	downFile := fmt.Sprintf("%s/%s_%s.down.sql", migrationsDir, timestamp, name)

	// Create empty migration files
	if err := os.WriteFile(upFile, []byte("-- Migration Up\n"), 0644); err != nil {
		log.Fatalf("Failed to create up migration file: %v", err)
	}
	if err := os.WriteFile(downFile, []byte("-- Migration Down\n"), 0644); err != nil {
		log.Fatalf("Failed to create down migration file: %v", err)
	}

	log.Printf("Migration files created:\n%s\n%s\n", upFile, downFile)
}
