package main

import (
	"log"
	"vstore/api"
	"vstore/database"
)

func main() {
	db, err := database.NewDatabase("mydb.data")

	if err != nil {
		log.Fatalf("Failed to create database: %v\n", err)
	}
	defer db.Close()

	api.StartServer(db)
}
