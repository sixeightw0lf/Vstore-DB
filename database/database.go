package database

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

var (
	// Hardcoded variable for the password, used for encryption and decryption
	dbPassword string = "password"
)

// Database struct includes an index map and a search index for tokenized search
type Database struct {
	mu            sync.RWMutex
	data          map[string]string
	index         map[string][]string        // Index based on a key characteristic
	searchIndex   map[string]map[string]bool // Inverted index for search: token -> map[key]bool
	file          string
	encryptionKey []byte
	gcm           cipher.AEAD
	connected     bool
}

func NewDatabase(filename string) (*Database, error) {
	db := &Database{
		data:        make(map[string]string),
		index:       make(map[string][]string),
		searchIndex: make(map[string]map[string]bool),
		file:        filename,
		connected:   false,
	}
	// Automatically set the encryption key using the hardcoded password
	if err := db.SetEncryptionKey(dbPassword); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *Database) SetEncryptionKey(key string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	hash := sha256.Sum256([]byte(key))
	db.encryptionKey = hash[:]

	block, err := aes.NewCipher(db.encryptionKey)
	if err != nil {
		return err
	}

	db.gcm, err = cipher.NewGCM(block)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) Connect() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.encryptionKey == nil {
		return fmt.Errorf("encryption key not set")
	}

	err := db.load()
	if err != nil {
		return err
	}

	db.connected = true
	return nil
}

func (db *Database) IsConnected() bool {
	return db.connected
}

func (db *Database) Disconnect() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if !db.connected {
		return fmt.Errorf("database not connected")
	}

	err := db.save()
	if err != nil {
		return err
	}

	db.data = make(map[string]string)
	db.index = make(map[string][]string)
	db.searchIndex = make(map[string]map[string]bool)
	db.connected = false
	return nil
}

func (db *Database) Set(key, value string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if !db.connected {
		return fmt.Errorf("database not connected")
	}

	// Update the actual data
	db.data[key] = value

	// Update the search index
	db.updateSearchIndex(key, value)

	return db.save()
}

func (db *Database) Get(key string) (string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if !db.connected {
		return "", fmt.Errorf("database not connected")
	}

	value, exists := db.data[key]
	if !exists {
		return "", os.ErrNotExist
	}

	return value, nil
}

// New function for getting data by ID and checking if pivotKey exists in value
func (db *Database) GetWithPivot(id, pivotKey string) (map[string]string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if !db.connected {
		return nil, fmt.Errorf("database not connected")
	}

	result := make(map[string]string)
	for key, value := range db.data {
		if key == id && strings.Contains(value, pivotKey) {
			result[key] = value
		}
	}

	if len(result) == 0 {
		return nil, os.ErrNotExist
	}

	return result, nil
}

func (db *Database) GetAll() map[string]string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if !db.connected {
		return nil
	}

	copy := make(map[string]string)
	for k, v := range db.data {
		copy[k] = v
	}
	fmt.Println("in get all data ", copy)
	return copy
}

func (db *Database) Delete(key string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if !db.connected {
		return fmt.Errorf("database not connected")
	}

	delete(db.data, key)
	return db.save()
}

func (db *Database) Search(keyword string) ([]string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if !db.connected {
		return nil, fmt.Errorf("database not connected")
	}

	var results []string
	for key, value := range db.data {
		if strings.Contains(value, keyword) {
			results = append(results, fmt.Sprintf("%s: %s", key, value))
		}
	}
	return results, nil
}

// New function to perform a fuzzy search
func (db *Database) SearchFuzzy(keyword string) ([]string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if !db.connected {
		return nil, fmt.Errorf("database not connected")
	}

	var results []string
	keywordLower := strings.ToLower(keyword)
	for key, value := range db.data {
		if strings.Contains(strings.ToLower(value), keywordLower) {
			results = append(results, fmt.Sprintf("%s: %s", key, value))
		}
	}
	return results, nil
}

// New function to query data with multiple terms
func (db *Database) Query(terms []string) ([]string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if !db.connected {
		return nil, fmt.Errorf("database not connected")
	}

	var results []string
	for key, value := range db.data {
		matchesAll := true
		for _, term := range terms {
			if !strings.Contains(value, term) {
				matchesAll = false
				break
			}
		}
		if matchesAll {
			results = append(results, fmt.Sprintf("%s: %s", key, value))
		}
	}
	return results, nil
}

// New function for fuzzy query with multiple terms
func (db *Database) QueryFuzzy(terms []string) ([]string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if !db.connected {
		return nil, fmt.Errorf("database not connected")
	}

	var results []string
	for key, value := range db.data {
		valueLower := strings.ToLower(value)
		matchesAll := true
		for _, term := range terms {
			termLower := strings.ToLower(term)
			if !strings.Contains(valueLower, termLower) {
				matchesAll = false
				break
			}
		}
		if matchesAll {
			results = append(results, fmt.Sprintf("%s: %s", key, value))
		}
	}
	return results, nil
}

func (db *Database) save() error {
	file, err := os.Create(db.file)
	if err != nil {
		return err
	}
	defer file.Close()

	var data bytes.Buffer
	encoder := gob.NewEncoder(&data)
	err = encoder.Encode(db.data)
	if err != nil {
		return err
	}

	compressedData, err := compress(data.Bytes())
	if err != nil {
		return err
	}

	encryptedData := db.encrypt(compressedData)

	_, err = file.Write(encryptedData)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) load() error {
	file, err := os.Open(db.file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No error if file doesn't exist
		}
		return err
	}
	defer file.Close()

	encryptedData, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	compressedData, err := db.decrypt(encryptedData)
	if err != nil {
		return err
	}

	data, err := decompress(compressedData)
	if err != nil {
		return err
	}

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&db.data)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) Close() error {
	return db.Disconnect()
}

// Helper function to update the search index when setting a new value
func (db *Database) updateSearchIndex(key, value string) {
	tokens := strings.Fields(value)
	for _, token := range tokens {
		if db.searchIndex[token] == nil {
			db.searchIndex[token] = make(map[string]bool)
		}
		db.searchIndex[token][key] = true
	}
}

func (db *Database) encrypt(data []byte) []byte {
	nonce := make([]byte, db.gcm.NonceSize())
	_, err := io.ReadFull(rand.Reader, nonce)
	if err != nil {
		panic(err.Error())
	}

	return db.gcm.Seal(nonce, nonce, data, nil)
}

func (db *Database) decrypt(data []byte) ([]byte, error) {
	nonceSize := db.gcm.NonceSize()
	nonce, encryptedData := data[:nonceSize], data[nonceSize:]

	return db.gcm.Open(nil, nonce, encryptedData, nil)
}

func compress(data []byte) ([]byte, error) {
	var compressedData bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedData)
	_, err := gzipWriter.Write(data)
	if err != nil {
		return nil, err
	}
	err = gzipWriter.Close()
	if err != nil {
		return nil, err
	}
	return compressedData.Bytes(), nil
}

func decompress(data []byte) ([]byte, error) {
	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()

	decompressedData, err := io.ReadAll(gzipReader)
	if err != nil {
		return nil, err
	}
	return decompressedData, nil
}
