package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"api-mia1/analizador"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/status", getStatus).Methods("GET")
	r.HandleFunc("/execute", getResponse).Methods("POST")

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173"},
		// AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)
	fmt.Println("Servidor escuchando en el puerto 5000")
	log.Fatal(http.ListenAndServe(":5000", handler))
}

func getStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, "La API está activa y funcionando correctamente!")
	fmt.Println("La API está activa y funcionando correctamente!")
}

type CodeRequest struct {
	Code string `json:"code"`
	Text string `json:"text"`
}

type CodeResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	Text    string `json:"text"`
}

type MultiCodeResponse struct {
	Results []CodeResponse `json:"results"`
}

func getResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Leer el cuerpo de la petición
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error al leer el cuerpo de la petición", http.StatusBadRequest)
		return
	}

	// Parsear el JSON recibido
	var codeReq CodeRequest
	err = json.Unmarshal(body, &codeReq)
	if err != nil {
		http.Error(w, "Error al parsear JSON", http.StatusBadRequest)
		return
	}

	// Separar por líneas y procesar cada comando
	lineas := strings.Split(codeReq.Code, "\n")
	var resultados []CodeResponse
	for _, linea := range lineas {
		linea = strings.TrimSpace(linea)
		if linea == "" {
			continue
		}
		comando, params := analizador.GetComandsParams(linea)
		resultados = append(resultados, CodeResponse{
			Message: "Comando recibido: " + comando,
			Code:    comando,
			Text:    params,
		})
		fmt.Printf("Comando recibido: %s, Parámetros: %s\n", comando, params)
	}

	// Crear la respuesta múltiple
	multiResp := MultiCodeResponse{Results: resultados}

	// Convertir la respuesta a JSON
	jsonResponse, err := json.Marshal(multiResp)
	if err != nil {
		http.Error(w, "Error al generar respuesta JSON", http.StatusInternalServerError)
		return
	}

	// Enviar la respuesta
	w.Write(jsonResponse)
}
