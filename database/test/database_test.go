package testDB

import (
	"os"
	"testing"
	"vstore/database"
)

func TestDatabaseSetAndGet(t *testing.T) {
	dbFilename := "testdb.data"
	db, err := database.NewDatabase(dbFilename)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer os.Remove(dbFilename) // Clean up after test

	testKey := "testKey"
	testValue := "testValue"

	// Test setting a value
	err = db.Set(testKey, testValue)
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	// Test getting the same value
	value, err := db.Get(testKey)
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}
	if value != testValue {
		t.Errorf("Expected value %s, got %s", testValue, value)
	}
}

func TestDatabasePersistence(t *testing.T) {
	dbFilename := "persistencetestdb.data"

	// Set a value in a new database instance
	{
		db, err := database.NewDatabase(dbFilename)
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		db.Set("persistKey", "persistValue")
		db.Close() // Close to ensure data is saved
	}

	// Create a new instance to test persistence
	{
		db, err := database.NewDatabase(dbFilename)
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		defer os.Remove(dbFilename) // Clean up after test

		value, err := db.Get("persistKey")
		if err != nil {
			t.Fatalf("Failed to get value: %v", err)
		}
		expectedValue := "persistValue"
		if value != expectedValue {
			t.Errorf("Expected value %s, got %s", expectedValue, value)
		}
	}
}
