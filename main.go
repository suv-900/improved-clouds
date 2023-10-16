package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/suv-900/blog/models"
	"github.com/suv-900/blog/routers"
)

// implement cors
func main() {

	router := mux.NewRouter()
	err := models.ConnectDB()
	if err != nil {
		log.Fatal(err)
		return
	}

	c := cors.New(cors.Options{
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowedOrigins:   []string{"http://localhost:3000"},
	})

	handler := c.Handler(router)
	routers.HandleRoutes(router)
	fmt.Println("Server started at port 8000")
	log.Fatal(http.ListenAndServe(":8000", handler))
}
