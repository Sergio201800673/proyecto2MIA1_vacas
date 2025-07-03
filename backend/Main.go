package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"api-mia1/analizador"
	diskmanager "api-mia1/diskManager"

	"api-mia1/structs"
	"encoding/binary"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/status", getStatus).Methods("GET")
	r.HandleFunc("/execute", getResponse).Methods("POST")
	r.HandleFunc("/confirm-delete", confirmDelete).Methods("POST")
	r.HandleFunc("/login", loginHandler).Methods("POST")
	r.HandleFunc("/discos", listarDiscosHandler).Methods("GET")
	r.HandleFunc("/particiones", particionesHandler).Methods("GET")

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

type LoginRequest struct {
	ID   string `json:"id"`
	User string `json:"user"`
	Pass string `json:"pass"`
}

type LoginResponse struct {
	Mensaje string `json:"mensaje"`
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}

	params := [][]string{
		{"-", "id", req.ID},
		{"-", "user", req.User},
		{"-", "pass", req.Pass},
	}
	result := diskmanager.Login(params)

	resp := LoginResponse{Mensaje: result}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type Disco struct {
	Nombre string `json:"nombre"`
}

func listarDiscosHandler(w http.ResponseWriter, r *http.Request) {
	discos := []Disco{}
	ruta := "/home/admin/sergio/backend/Discos/"

	files, err := os.ReadDir(ruta)
	if err != nil {
		http.Error(w, "No se pudieron leer los discos", http.StatusInternalServerError)
		return
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".dsk" {
			discos = append(discos, Disco{Nombre: file.Name()})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(discos)
}

type Particion struct {
	Nombre string `json:"nombre"`
	Tipo   string `json:"tipo"`
	Tamano int64  `json:"tamano"`
}

func particionesHandler(w http.ResponseWriter, r *http.Request) {
	nombreDisco := r.URL.Query().Get("disco")
	if nombreDisco == "" {
		http.Error(w, "Falta el parámetro 'disco'", http.StatusBadRequest)
		return
	}

	ruta := filepath.Join("/home/admin/sergio/backend/Discos/", nombreDisco)
	file, err := os.Open(ruta)
	if err != nil {
		http.Error(w, "No se pudo abrir el disco", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	var mbr structs.MBR
	binary.Read(file, binary.LittleEndian, &mbr)

	particiones := []Particion{}
	for _, p := range mbr.Partitions {
		if p.PartStatus != [1]byte{0} { // Solo particiones activas
			particiones = append(particiones, Particion{
				Nombre: string(p.PartName[:]),
				Tipo:   string(p.PartType[:]),
				Tamano: int64(p.PartSize),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(particiones)
}
