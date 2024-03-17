package database

import (
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"vstore/database/query"
)

// Database struct includes an index map and a search index for tokenized search
type Database struct {
	mu          sync.RWMutex
	data        map[string]string
	index       map[string][]string        // Index based on a key characteristic
	searchIndex map[string]map[string]bool // Inverted index for search: token -> map[key]bool
	file        string
}

func NewDatabase(filename string) (*Database, error) {
	db := &Database{
		data:        make(map[string]string),
		index:       make(map[string][]string),
		searchIndex: make(map[string]map[string]bool),
		file:        filename,
	}
	err := db.load()
	if err != nil {
		return nil, err
	}
	return db, nil
}

// extractIndexKey extracts an index key from the given key.
func extractIndexKey(key string) string {
	if len(key) > 0 {
		return string(key[0])
	}
	return ""
}

// Now include updateSearchIndex call in Set method
func (db *Database) Set(key, value string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Update the actual data
	db.data[key] = value

	// Update the search index
	db.updateSearchIndex(key, value)

	return db.save()
}

func (db *Database) GetAll() map[string]string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// Return a copy of the data to avoid external modification
	dataCopy := make(map[string]string)
	for k, v := range db.data {
		dataCopy[k] = v
	}
	return dataCopy
}

// Update a value for a given key
func (db *Database) Update(key, newValue string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.data[key]; !exists {
		return fmt.Errorf("key not found: %s", key)
	}

	db.data[key] = newValue
	return db.save()
}

// Delete a key-value pair
func (db *Database) Delete(key string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.data[key]; !exists {
		return fmt.Errorf("key not found: %s", key)
	}

	delete(db.data, key)
	return db.save()
}

// Example query execution function
func (db *Database) ExecuteQuery(queryString string) (string, error) {
	// Using `query.` prefix to reference ParseQuery
	queryObj, err := query.ParseQuery(queryString)
	if err != nil {
		return "", err
	}

	switch queryObj.Type {
	// Using `query.` prefix to reference QueryInsert and QuerySelect
	case query.QueryInsert:
		return "", db.Set(queryObj.Key, queryObj.Value)
	case query.QuerySelect:
		value, err := db.Get(queryObj.Key)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Value: %s", value), nil
	default:
		return "", errors.New("unsupported query type")
	}
}

func (db *Database) Get(key string) (string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	value, exists := db.data[key]
	if !exists {
		return "", os.ErrNotExist
	}

	return value, nil
}

func (db *Database) load() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	file, err := os.Open(db.file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&db.data)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) save() error {
	file, err := os.Create(db.file)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(db.data)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) Close() error {
	return db.save()
}

// Tokenize a value into words (simple approach)
func tokenize(value string) []string {
	// This is a very basic tokenizer; consider using a more sophisticated approach for production
	return strings.Fields(value)
}

// Update the search index when setting a new value
func (db *Database) updateSearchIndex(key, value string) {
	tokens := tokenize(value)
	for _, token := range tokens {
		if db.searchIndex[token] == nil {
			db.searchIndex[token] = make(map[string]bool)
		}
		db.searchIndex[token][key] = true
	}
}

// Search by keyword
func (db *Database) SearchKeyword(keyword string) []string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	keys := make([]string, 0)
	if entries, ok := db.searchIndex[keyword]; ok {
		for key := range entries {
			keys = append(keys, key)
		}
	}
	return keys
}

func (db *Database) Search(query string) ([]string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []string
	isRegex := strings.Contains(query, ".*") // Simplistic check for regex pattern
	if isRegex {
		regex, err := regexp.Compile(query)
		if err != nil {
			return nil, err
		}
		for key, value := range db.data {
			if regex.MatchString(value) {
				results = append(results, fmt.Sprintf("%s: %s", key, value))
			}
		}
	} else {
		for key, value := range db.data {
			if strings.Contains(value, query) {
				results = append(results, fmt.Sprintf("%s: %s", key, value))
			}
		}
	}
	return results, nil
}
