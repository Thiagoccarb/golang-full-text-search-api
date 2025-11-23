package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	"go.elastic.co/apm/module/apmgorilla/v2"
	"go.elastic.co/apm/module/apmsql/v2"
	_ "go.elastic.co/apm/module/apmsql/v2/pq"
	"go.elastic.co/apm/v2"
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
	ES *elasticsearch.Client
}

func main() {
	apm.DefaultTracer()

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := apmsql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	log.Println("Connected to database successfully")

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{os.Getenv("ELASTICSEARCH_URL")},
	})
	if err != nil {
		log.Fatal("Failed to create Elasticsearch client:", err)
	} else {
		log.Println("Connected to Elasticsearch successfully")
	}

	app := &App{DB: db, ES: es}

	if app.ES != nil {
		go app.syncData()
	}

	router := mux.NewRouter()
	router.Use(apmgorilla.Middleware())
	router.HandleFunc("/search", app.searchHandler).Methods("GET")
	router.HandleFunc("/search/optimized", app.optimizedSearchHandler).Methods("GET")
	router.HandleFunc("/health", healthHandler).Methods("GET")

	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func (app *App) syncData() {
	log.Println("Syncing data to Elasticsearch...")

	mapping := `{"mappings":{"properties":{"id":{"type":"integer"},"text":{"type":"text"}}}}`
	res, err := app.ES.Indices.Create("search_data", app.ES.Indices.Create.WithBody(strings.NewReader(mapping)))
	if err != nil {
		log.Printf("Error creating index: %v", err)
		return
	}
	if res.IsError() && !strings.Contains(res.String(), "resource_already_exists_exception") {
		log.Printf("Index creation failed: %s", res.String())
	}
	res.Body.Close()

	rows, err := app.DB.Query("SELECT id, text FROM search_data")
	if err != nil {
		log.Printf("Sync error: %v", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var result SearchResult
		if err := rows.Scan(&result.ID, &result.Text); err != nil {
			continue
		}

		doc, _ := json.Marshal(result)
		app.ES.Index("search_data", strings.NewReader(string(doc)), app.ES.Index.WithDocumentID(fmt.Sprintf("%d", result.ID)))

		count++
		if count%5000 == 0 {
			log.Printf("Synced %d records", count)
		}
	}

	log.Printf("Sync complete: %d records indexed to Elasticsearch", count)
}

func (app *App) searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Missing 'query' parameter", http.StatusBadRequest)
		return
	}

	searchPattern := "%" + query + "%"
	searchQuery := "SELECT id, text FROM search_data WHERE text ILIKE $1"

	rows, err := app.DB.QueryContext(r.Context(), searchQuery, searchPattern)
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
			continue
		}
		results = append(results, result)
	}

	response := SearchResponse{
		Results: results,
		Query:   query,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (app *App) optimizedSearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Missing 'query' parameter", http.StatusBadRequest)
		return
	}

	if results := app.searchElasticsearch(r.Context(), query); results != nil {
		response := SearchResponse{
			Results: *results,
			Query:   query,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"error": "Elasticsearch search failed"})
}

func (app *App) searchElasticsearch(ctx context.Context, query string) *[]SearchResult {
	span, ctx := apm.StartSpan(ctx, "elasticsearch_search", "db.elasticsearch")
	defer span.End()

	searchQuery := fmt.Sprintf(`{
        "query": {
            "match": {
                "text": "%s"
            }
        },
        "size": 100
    }`, query)

	res, err := app.ES.Search(
		app.ES.Search.WithIndex("search_data"),
		app.ES.Search.WithBody(strings.NewReader(searchQuery)),
		app.ES.Search.WithContext(ctx),
	)

	if err != nil || res.IsError() {
		return nil
	}
	defer res.Body.Close()

	var searchResult struct {
		Hits struct {
			Hits []struct {
				Source SearchResult `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil
	}

	var results []SearchResult
	for _, hit := range searchResult.Hits.Hits {
		results = append(results, hit.Source)
	}

	return &results
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
