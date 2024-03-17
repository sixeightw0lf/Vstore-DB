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
		case http.MethodDelete:
			if len(parts) == 2 && parts[0] == "data" {
				h.handleDeleteRecord(w, r, parts[1])
			} else {
				http.NotFound(w, r)
			}
		case "password":
			h.handleGetEncryptionKey(w, r)
		default:
			http.NotFound(w, r)
		}
	case http.MethodPost:
		switch parts[0] {
		case "data":
			h.handlePostData(w, r)
		case "password":
			h.handleSetEncryptionKey(w, r)
		case "connect":
			h.handleConnect(w, r)
		case "disconnect":
			h.handleDisconnect(w, r)
		default:
			http.NotFound(w, r)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *APIHandler) handlePostData(w http.ResponseWriter, r *http.Request) {
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

func (h *APIHandler) handleGetAll(w http.ResponseWriter, _ *http.Request) {
	data := h.db.GetAll()
	if data == nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Database not connected")
		return
	}
	fmt.Println("get all data:", data)
	h.respondWithJSON(w, data)
}

func (h *APIHandler) handleGetByID(w http.ResponseWriter, r *http.Request, id string) {
	data, err := h.db.Get(id)
	if err != nil {
		if err.Error() == "database not connected" {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "Database not connected")
		} else {
			http.NotFound(w, r)
		}
		return
	}
	h.respondWithJSON(w, map[string]string{id: data})
}

func (h *APIHandler) handleGetByIDAndPivotKey(w http.ResponseWriter, r *http.Request, id, pivotKey string) {
	data, err := h.db.GetWithPivot(id, pivotKey)
	if err != nil {
		if err.Error() == "database not connected" {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "Database not connected")
		} else {
			http.NotFound(w, r)
		}
		return
	}
	fmt.Println("get by id and pivot key:", data)
	h.respondWithJSON(w, data)
}

func (h *APIHandler) handleSearch(w http.ResponseWriter, _ *http.Request, keyword string, fuzzy bool) {
	var results []string
	var err error
	if fuzzy {
		results, err = h.db.SearchFuzzy(keyword)
	} else {
		results, err = h.db.Search(keyword)
	}
	if err != nil {
		if err.Error() == "database not connected" {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "Database not connected")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Search error: %v", err)
		}
		return
	}
	fmt.Println("search results:", results)
	h.respondWithJSON(w, results)
}

func (h *APIHandler) handleQuery(w http.ResponseWriter, _ *http.Request, terms []string, fuzzy bool) {
	var results []string
	var err error
	if fuzzy {
		results, err = h.db.QueryFuzzy(terms)
	} else {
		results, err = h.db.Query(terms)
	}
	if err != nil {
		if err.Error() == "database not connected" {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "Database not connected")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Query error: %v", err)
		}
		return
	}
	fmt.Println("query results:", results)
	h.respondWithJSON(w, results)
}

func (h *APIHandler) handleDeleteRecord(w http.ResponseWriter, _ *http.Request, id string) {
	err := h.db.Delete(id)
	if err != nil {
		if err.Error() == "database not connected" {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "Database not connected")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Delete error: %v", err)
		}
		return
	}
	fmt.Println("record deleted successfully", id)
	h.respondWithJSON(w, map[string]string{"message": "Record deleted successfully"})
}

func (h *APIHandler) handleSetEncryptionKey(w http.ResponseWriter, r *http.Request) {
	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Failed to decode request body: %v", err)
		return
	}
	password, ok := data["password"]
	fmt.Println("password set.")
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Missing password field")
		return
	}

	err = h.db.SetEncryptionKey(password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to set encryption key: %v", err)
		return
	}

	h.respondWithJSON(w, map[string]string{"message": "Encryption key set successfully"})
}

func (h *APIHandler) handleGetEncryptionKey(w http.ResponseWriter, _ *http.Request) {
	if h.db.IsConnected() {
		h.respondWithJSON(w, map[string]string{"message": "Database is connected"})
	} else {
		h.respondWithJSON(w, map[string]string{"message": "Database is not connected"})
	}
}

func (h *APIHandler) handleConnect(w http.ResponseWriter, _ *http.Request) {
	err := h.db.Connect()
	fmt.Println("client trying to connect....")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(w, "Failed to connect to database: %v", err)
		return
	}
	fmt.Println("database connected successfully")
	h.respondWithJSON(w, map[string]string{"message": "Database connected successfully"})
}

func (h *APIHandler) handleDisconnect(w http.ResponseWriter, _ *http.Request) {
	err := h.db.Disconnect()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to disconnect from database: %v", err)
		return
	}
	fmt.Println("database disconnected successfully")
	h.respondWithJSON(w, map[string]string{"message": "Database disconnected successfully"})
}

func (h *APIHandler) respondWithJSON(w http.ResponseWriter, data interface{}) {
	response, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to encode response: %v", err)
		return
	}
	fmt.Println("DB response:", data)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func StartServer(db *database.Database) {
	apiHandler := NewAPIHandler(db)
	http.Handle("/", apiHandler)
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
