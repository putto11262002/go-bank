package main

import (
	"encoding/json"
	"log"
	"net/http"
)


func main()  {
	connStr := "user=postgres dbname=postgres password=password sslmode=disable"
	store, err := NewPostgresStore(connStr)
	if err != nil {
		log.Fatal(err)
	}
	if err := store.init(); err != nil {
		log.Fatal(err)
	}

	server := &APIServer{listenAddr: ":3000", store: store}
	server.Run()
}


func WriteJSON(w http.ResponseWriter, status int, val any) error {
	w.Header().Set("content-Type", "application/json") 
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(val)
}