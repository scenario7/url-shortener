package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

type URLObject struct {
	gorm.Model
	Key       string `json:"key"`
	Original  string `json:"original"`
	Shortened string `json:"shortened"`
}

var db *gorm.DB
var err error

func main() {
	router := mux.NewRouter()
	db, err = gorm.Open("postgres", "host=localhost port=5432 sslmode=disable dbname=test")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.AutoMigrate(&URLObject{})

	router.HandleFunc("/urls", addUrl).Methods("POST")
	router.HandleFunc("/urls", getUrls).Methods("GET")
	router.HandleFunc("/urls/{key}", getUrl).Methods("GET")

	log.Fatal(http.ListenAndServe(":8000", router))
	fmt.Println("Starting Server")

}

func addUrl(w http.ResponseWriter, r *http.Request) {
	var url URLObject
	err := json.NewDecoder(r.Body).Decode(&url)
	if err != nil {
		log.Fatal(err)
	}
	rand.Seed(time.Now().UnixNano())
	charset := "abcdefghijklmnopqrstuvwxyz"
	shortened := make([]byte, 6)
	for i := range shortened {
		shortened[i] = charset[rand.Intn(len(charset))]
	}

	url.Shortened = string(shortened)
	db.Create(&url)
	json.NewEncoder(w).Encode("Your shortened URL is " + url.Shortened)
}

func getUrls(w http.ResponseWriter, r *http.Request) {
	var urls []URLObject

	db.Find(&urls)

	json.NewEncoder(w).Encode(urls)
}

func getUrl(w http.ResponseWriter, r *http.Request) {
	var url URLObject
	vars := mux.Vars(r)
	key := vars["key"]
	if err := db.Where("key = ?", key).First(&url).Error; err != nil {
		if err := db.Where("shortened = ?", key).First(&url).Error; err != nil {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}
	}

	json.NewEncoder(w).Encode(url.Original)
}
