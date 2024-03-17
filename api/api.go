package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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

// ServeHTTP handles the HTTP requests
func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	switch r.Method {
	case http.MethodGet:
		switch parts[0] {
		case "get":
			if len(parts) == 2 {
				if parts[1] == "all" {
					h.handleGetAll(w, r)
				} else {
					h.handleGetByID(w, r, parts[1])
				}
			} else if len(parts) == 3 {
				h.handleGetByIDAndPivotKey(w, r, parts[1], parts[2])
			}
		case "search":
			if len(parts) == 2 {
				h.handleSearch(w, r, parts[1], false)
			} else if len(parts) == 3 && parts[2] == "fuzzy" {
				h.handleSearch(w, r, parts[1], true)
			}
		case "query":
			h.handleQuery(w, r, parts[1:], strings.Contains(r.URL.RawQuery, "fuzzy=true"))
		default:
			http.NotFound(w, r)
		}
	case http.MethodPost:
		h.handlePost(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *APIHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Failed to decode request body: %v", err)
		return
	}

	for key, value := range data {
		fmt.Print("inserting::::", "key: ", key, " value: ", value, "\n")
		err := h.db.Set(key, value)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Failed to set value: %v", err)
			return
		}
	}

	h.respondWithJSON(w, map[string]string{"message": "Data stored successfully"})
}

func (h *APIHandler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	data := h.db.GetAll()
	print("get all data:", data, "\n")
	h.respondWithJSON(w, data)
}

func (h *APIHandler) handleGetByID(w http.ResponseWriter, r *http.Request, id string) {
	data, err := h.db.Get(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	h.respondWithJSON(w, map[string]string{id: data})
}

func (h *APIHandler) handleGetByIDAndPivotKey(w http.ResponseWriter, r *http.Request, id, pivotKey string) {
	data, err := h.db.GetWithPivot(id, pivotKey)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	h.respondWithJSON(w, data)
}

func (h *APIHandler) handleSearch(w http.ResponseWriter, r *http.Request, keyword string, fuzzy bool) {
	var results []string
	var err error
	if fuzzy {
		results, err = h.db.SearchFuzzy(keyword)
	} else {
		results, err = h.db.Search(keyword)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Search error: %v", err)
		return
	}
	fmt.Print("search results:", results, "\n")
	h.respondWithJSON(w, results)
}

func (h *APIHandler) handleQuery(w http.ResponseWriter, r *http.Request, terms []string, fuzzy bool) {
	var results []string
	var err error
	if fuzzy {
		results, err = h.db.QueryFuzzy(terms)
	} else {
		results, err = h.db.Query(terms)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Query error: %v", err)
		return
	}
	h.respondWithJSON(w, results)
}

func (h *APIHandler) respondWithJSON(w http.ResponseWriter, data interface{}) {
	response, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to encode response: %v", err)
		return
	}
	fmt.Print("response:", data, "\n")
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func StartServer(db *database.Database) {
	apiHandler := NewAPIHandler(db)
	http.Handle("/", apiHandler)
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
