package db

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"time"
)

func Establish_db_conn() *sql.DB {
	// Load .env file according to app env
	var app_env = os.Getenv("APP_ENV")
	if app_env == "" {
		app_env = "dev"
	}
	err := godotenv.Load("../" + app_env + ".env")
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
	var db *sql.DB
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("Connection could not be opened. Error:", err)
	}
	log.Print("Successfully conected to database!")

	// Return db object
	return db
}

func InsertOrUpdateEtf(db *sql.DB, id string, name string, fundVolume string, isDistributing bool, releaseDate time.Time, replicationMethod string, shareClassVolume string, totalExpenseRatio float32) {
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
