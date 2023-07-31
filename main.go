package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
)

// Document represents the structure of a document.
type Document struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

var redisClient *redis.Client

func initRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Replace this with your Redis server address
		DB:   0,
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
}

// StoreDocumentHandler handles storing a new document in Redis.
func StoreDocumentHandler(w http.ResponseWriter, r *http.Request) {
	var document Document

	err := json.NewDecoder(r.Body).Decode(&document)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if document.ID == "" {
		http.Error(w, "Document ID must not be empty", http.StatusBadRequest)
		return
	}

	jsonData, err := json.Marshal(document)
	if err != nil {
		http.Error(w, "Error converting document to JSON", http.StatusInternalServerError)
		return
	}

	err = redisClient.Set(document.ID, jsonData, 0).Err()
	if err != nil {
		http.Error(w, "Failed to store document in Redis", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetDocumentHandler retrieves a document from Redis based on the provided ID.
func GetDocumentHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	docID := params["id"]

	val, err := redisClient.Get(docID).Result()
	if err == redis.Nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to retrieve document from Redis", http.StatusInternalServerError)
		return
	}

	var document Document
	err = json.Unmarshal([]byte(val), &document)
	if err != nil {
		http.Error(w, "Error unmarshalling document from JSON", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(document)
}

func sayHello(name string) string {
	return fmt.Sprintf("Hello %s", name)
}

func main() {
	msg := sayHello("Alicea")
	fmt.Println(msg)

	initRedis()

	router := mux.NewRouter()
	router.HandleFunc("/documents", StoreDocumentHandler).Methods("POST")
	router.HandleFunc("/documents/{id}", GetDocumentHandler).Methods("GET")

	port := 8080
	fmt.Printf("Server listening on port %d...\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
}
