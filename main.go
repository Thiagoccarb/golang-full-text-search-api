package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type SearchResult struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Query   string         `json:"query"`
}

type App struct {
	DB *sql.DB
}

func main() {
	db, err := sql.Open("postgres", "host=postgres port=5432 user=postgres password=password dbname=searchdb sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Connected to database successfully")

	app := &App{DB: db}

	// Setup routes
	router := mux.NewRouter()
	router.HandleFunc("/search", app.searchHandler).Methods("GET")
	router.HandleFunc("/health", healthHandler).Methods("GET")

	// Start server
	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func (app *App) searchHandler(w http.ResponseWriter, r *http.Request) {
	// Get query parameter
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Missing 'query' parameter", http.StatusBadRequest)
		return
	}

	// Simple search using ILIKE - no pagination, no count
	searchPattern := "%" + query + "%"
	searchQuery := "SELECT id, text FROM search_data WHERE text ILIKE $1"

	rows, err := app.DB.Query(searchQuery, searchPattern)
	if err != nil {
		log.Printf("Error executing search: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		if err := rows.Scan(&result.ID, &result.Text); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Create simple response
	response := SearchResponse{
		Results: results,
		Query:   query,
	}

	// Set headers and encode JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "JSON encoding error", http.StatusInternalServerError)
		return
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
