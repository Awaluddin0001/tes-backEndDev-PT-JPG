package main

import (
	"log"
	"net/http"

	beHandlers "backend/server/handlers"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	// Inisialisasi database
	err := beHandlers.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	// Definisikan route
	r.HandleFunc("/user", beHandlers.RegisterHandler).Methods("POST")
	r.HandleFunc("/login", beHandlers.LoginHandler).Methods("POST")
	r.HandleFunc("/refresh-token", beHandlers.RefreshTokenHandler).Methods("POST")
	r.HandleFunc("/sales", beHandlers.InputSalesHandler).Methods("POST")
	r.HandleFunc("/report", beHandlers.ReportSalesHandler).Methods("POST")

	log.Fatal(http.ListenAndServe(":3000", cors(r)))
}
