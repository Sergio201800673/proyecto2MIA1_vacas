package backend

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/", initAPi).Methods("GET")

}

func initApi(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Backend server is starting...")
}
