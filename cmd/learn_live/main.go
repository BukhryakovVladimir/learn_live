package main

import (
	"fmt"
	"github.com/BukhryakovVladimir/learn_live/internal/handlers/learn_live_handler"
	"github.com/BukhryakovVladimir/learn_live/internal/routes"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"os"
)

func main() {
	err := routes.InitConnPool()

	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	r := chi.NewRouter()

	learn_live_handler.SetupRoutes(r)

	strPort := os.Getenv("PORT")
	if strPort == "" {
		log.Fatalf("Environment variable PORT is empty.")
	}
	port := fmt.Sprintf(":%s", strPort)

	err = http.ListenAndServe(port, r)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
