package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/status", getStatus).Methods("GET")
	r.HandleFunc("/response", getResponse).Methods("GET")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)
	fmt.Println("Servidor escuchando en el puerto 8080")
	log.Fatal(http.ListenAndServe(":8080", handler))

}

func getStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, "La API está activa y funcionando correctamente!")
	fmt.Println("La API está activa y funcionando correctamente!")
}

func getResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := `{"message": "Hello, World!"}`
	fmt.Fprint(w, response)
	fmt.Println("Response sent successfully!")
}
