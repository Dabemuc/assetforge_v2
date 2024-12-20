package main

import (
	"backend/db"
	"backend/scraper"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	log.Println("Starting assertforge_v2 backend ...")

	// Establish connection to db
	db.Establish_db_conn()

	//id := "ie00b579f325"

	scraper.ScrapeEtf(nil)
	os.Exit(0)

	// Start Server
	start_server()
}

func start_server() {
	// Serve webpage
	http.HandleFunc("/", serveRoot)

	// Serve api endpoints
	http.HandleFunc("/api/fetchEtfProfile", fetchEtfProfile)

	// Start the server
	port := ":8080"
	log.Print("Server started at http://localhost:", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func serveRoot(w http.ResponseWriter, r *http.Request) {
	log.Print("Received request to serveRoot")
	http.ServeFile(w, r, "./frontend/dist/index.html")
}

func fetchEtfProfile(w http.ResponseWriter, r *http.Request) {
	log.Print("Received request to fetchEtfProfile with params: ", r.URL.Query())
	var symbol = r.URL.Query().Get("symbol")
	fmt.Fprintf(w, `{"symbol": "%s"}`, symbol)
}
