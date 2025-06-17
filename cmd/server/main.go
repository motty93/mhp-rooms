package main

import (
	"fmt"
	"log"
	"net/http"

	"mhp-rooms/internal/handlers"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", handlers.HomeHandler).Methods("GET")
	r.HandleFunc("/hello", handlers.HelloHandler).Methods("GET")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	fmt.Println("サーバーを起動しています... http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
