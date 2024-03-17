package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"vstore/database"
)

// APIHandler struct holds a reference to the database
type APIHandler struct {
	db *database.Database
}

// NewAPIHandler creates a new APIHandler
func NewAPIHandler(db *database.Database) *APIHandler {
	return &APIHandler{db: db}
}

// ServeHTTP handles the HTTP requests for CRUD operations and search
func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost: // Create
		h.handleCreate(w, r)
	case http.MethodGet: // Read or Search
		query := r.URL.Query().Get("query")
		if query != "" {
			h.handleSearch(w, r, query) // Handle search if a query parameter is present
		} else {
			h.handleRead(w, r) // Handle regular read if no query parameter is present
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Unsupported method")
	}
}

// handleSearch processes search requests
func (h *APIHandler) handleSearch(w http.ResponseWriter, r *http.Request, query string) {
	results, err := h.db.Search(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error searching: %v", err)
		return
	}
	jsonResponse, err := json.Marshal(results)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error encoding search results: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

// handleCreate processes create (insert) requests
func (h *APIHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error parsing request body: %v", err)
		return
	}
	err = h.db.Set(data.Key, data.Value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error setting value: %v", err)
		return
	}
	fmt.Fprintf(w, "Value set successfully")
}

// handleRead processes read (select) requests
func (h *APIHandler) handleRead(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Key is required")
		return
	}
	value, err := h.db.Get(key)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Key not found")
		return
	}
	fmt.Fprintf(w, "Value: %s", value)
}

func StartServer(db *database.Database) {
	apiHandler := NewAPIHandler(db)
	http.Handle("/", apiHandler)
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
