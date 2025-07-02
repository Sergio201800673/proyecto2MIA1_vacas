package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"api-mia1/analizador"
	diskmanager "api-mia1/diskManager"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/status", getStatus).Methods("GET")
	r.HandleFunc("/execute", getResponse).Methods("POST")
	r.HandleFunc("/confirm-delete", confirmDelete).Methods("POST")

	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:5173",
			"http://proyecto2-archivos-srmdb.s3-website.us-east-2.amazonaws.com",
		},
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
	Output  string         `json:"output"`
}

func getResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

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
	var outputMsg strings.Builder

	for _, linea := range lineas {
		linea = strings.TrimSpace(linea)

		if linea == "" {
			continue
		}

		comando, params := analizador.GetComandsParams(linea)

		// Llamar al analizador de comandos para mostrar el resultado
		output := analizador.AnalizerCommand(comando, params)
		fmt.Println(output)

		// Agregar la salida del comando al mensaje
		if output != "" {
			outputMsg.WriteString(output)
			outputMsg.WriteString("\n")
		}

		resultados = append(resultados, CodeResponse{
			Message: "Comando recibido: " + comando,
			Code:    comando,
			Text:    params,
		})
		/* fmt.Printf("Comando recibido: %s, Parámetros: %s\n", comando, params) */
	}

	// Si no hay salida de comandos, crear mensaje por defecto
	if outputMsg.Len() == 0 {
		outputMsg.WriteString("Comandos procesados:\n")
		for _, resultado := range resultados {
			outputMsg.WriteString(fmt.Sprintf("- %s: %s\n", resultado.Code, resultado.Text))
		}
	}

	// Crear la respuesta múltiple
	multiResp := MultiCodeResponse{
		Results: resultados,
		Output:  outputMsg.String(),
	}

	// Convertir la respuesta a JSON
	jsonResponse, err := json.Marshal(multiResp)
	if err != nil {
		http.Error(w, "Error al generar respuesta JSON", http.StatusInternalServerError)
		return
	}

	// Enviar la respuesta con codificación UTF-8
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(jsonResponse)))
	w.Write(jsonResponse)
}

type ConfirmDeleteRequest struct {
	Filename string `json:"filename"`
	FullPath string `json:"fullPath"`
}

func confirmDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Leer el cuerpo de la petición
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error al leer el cuerpo de la petición", http.StatusBadRequest)
		return
	}

	// Parsear el JSON recibido
	var confirmReq ConfirmDeleteRequest
	err = json.Unmarshal(body, &confirmReq)
	if err != nil {
		http.Error(w, "Error al parsear JSON", http.StatusBadRequest)
		return
	}

	// Confirmar la eliminación
	result := diskmanager.ConfirmRmdisk(confirmReq.Filename, confirmReq.FullPath)

	// Crear la respuesta
	response := CodeResponse{
		Message: result,
		Code:    "confirm-delete",
		Text:    fmt.Sprintf("filename: %s, fullPath: %s", confirmReq.Filename, confirmReq.FullPath),
	}

	// Convertir la respuesta a JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error al generar respuesta JSON", http.StatusInternalServerError)
		return
	}

	// Enviar la respuesta
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(jsonResponse)))
	w.Write(jsonResponse)
}
