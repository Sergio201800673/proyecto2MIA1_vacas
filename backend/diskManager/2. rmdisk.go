package diskmanager

import (
	"fmt"
	"os"
	"strings"
)

func Rmdisk(params [][]string) string {
	var letra string = ""

	for _, param := range params {
		key := strings.ToLower(param[1])
		val := strings.Trim(param[2], "\"")

		if key == "driveletter" {
			letra = strings.ToUpper(val)
		} else {
			return "\n⚠️ Parámetro no reconocido:" + key
		}
	}
	if letra == "" {
		return "\n❌ Error: el parámetro -driveletter es obligatorio."
	}

	filename := letra + ".dsk"
	fullPath := rutaBase + filename
	fmt.Println(fullPath)

	// Verificar existencia
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "❌ Error: el disco " + filename + " no existe."
	}

	// En lugar de pedir confirmación directamente, retornamos un mensaje especial
	// que el frontend puede interpretar para mostrar un diálogo de confirmación
	return "\nCONFIRM_DELETE:" + filename + ":" + fullPath
}

// Nueva función para confirmar la eliminación
func ConfirmRmdisk(filename string, fullPath string) string {
	// Eliminar archivo
	err := os.Remove(fullPath)
	if err != nil {
		return "\n❌ Error al eliminar el archivo:" + err.Error()
	} else {
		return "\n✅ Disco " + filename + " eliminado exitosamente."
	}
}
