package diskmanager

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func Rmdisk(params [][]string) {
	var letra string = ""

	for _, param := range params {
		key := strings.ToLower(param[1])
		val := strings.Trim(param[2], "\"")

		if key == "driveletter" {
			letra = strings.ToUpper(val)
		} else {
			fmt.Println("⚠️ Parámetro no reconocido:", key)
		}
	}
	if letra == "" {
		fmt.Println("❌ Error: el parámetro -driveletter es obligatorio.")
		return
	}

	filename := letra + ".dsk"
	fullPath := rutaBase + filename

	// Verificar existencia
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		fmt.Println("❌ Error: el disco", filename, "no existe.")
		return
	}

	// Confirmación del usuario
	fmt.Println("⚠️ ¿Está seguro que desea eliminar el disco", filename, "? (s/n): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	if input != "s" {
		fmt.Println("❌ Operación cancelada.")
		return
	}

	// Eliminar archivo
	err := os.Remove(fullPath)
	if err != nil {
		fmt.Println("❌ Error al eliminar el archivo:", err)
	} else {
		fmt.Println("✅ Disco", filename, "eliminado exitosamente.")
	}
}
