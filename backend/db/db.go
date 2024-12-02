package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"sync"
	"time"
)

var (
	db   *sql.DB
	once sync.Once
)

func Establish_db_conn() {
	once.Do(func() {
		// Load .env file according to app env
		var app_env = os.Getenv("APP_ENV")
		if app_env == "" {
			app_env = "dev"
		}
		_, filename, _, _ := runtime.Caller(0)                           // Gets this files path
		projectRoot := filepath.Join(filepath.Dir(filename), "..", "..") // Goes up two levels to root dir
		envFile := filepath.Join(projectRoot, "backend", app_env+".env")
		err := godotenv.Load(envFile)
		if err != nil {
			log.Fatal("Error loading .env file:", err)
		}

		// Read env vars
		DB_USER := os.Getenv("ASSETFORGE_V2_DB_USER")
		DB_PASSWORD := os.Getenv("ASSETFORGE_V2_DB_PASSWORD")
		DB_NAME := os.Getenv("ASSETFORGE_V2_DB_NAME")
		DB_HOST := os.Getenv("ASSETFORGE_V2_DB_HOST")
		DB_PORT := os.Getenv("ASSETFORGE_V2_DB_PORT")

		// Connect to db
		connStr := "user=" + DB_USER + " password=" + DB_PASSWORD + " dbname=" + DB_NAME + " sslmode=disable host=" + DB_HOST + " port=" + DB_PORT
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Fatal("Error connecting to the database:", err)
		}
		err = db.Ping()
		if err != nil {
			log.Fatal("Connection could not be opened. Error:", err)
		}
		log.Print("Successfully conected to database!")
	})
}

func GetDb() *sql.DB {
	if db == nil {
		log.Fatal("db.Getdb() called without connection being established first. First call db.Establish_db_conn().")
	}
	return db
}

func InsertOrUpdateEtf(id string, name string, fundVolume string, isDistributing bool, releaseDate time.Time, replicationMethod string, shareClassVolume string, totalExpenseRatio float32) {
	// Ensure releaseDate is only a date, not a timestamp.
	releaseDate = releaseDate.Truncate(24 * time.Hour)

	var queryString = `
		INSERT INTO t_etf (id, name, fundVolume, isDistributing, releaseDate, replicationMethod, shareClassVolume, totalExpenseRatio)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id)
		DO UPDATE SET
			name = EXCLUDED.name,
			fundVolume = EXCLUDED.fundVolume,
			isDistributing = EXCLUDED.isDistributing,
			releaseDate = EXCLUDED.releaseDate,
			replicationMethod = EXCLUDED.replicationMethod,
			shareClassVolume = EXCLUDED.shareClassVolume,
			totalExpenseRatio = EXCLUDED.totalExpenseRatio`

	_, err := db.Exec(queryString, id, name, fundVolume, isDistributing, releaseDate, replicationMethod, shareClassVolume, totalExpenseRatio)
	if err != nil {
		log.Printf("Error executing query: %v", err)
	}
}
