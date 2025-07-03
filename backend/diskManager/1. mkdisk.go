package diskmanager

import (
	"api-mia1/structs"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var rutaBase = "/home/admin/sergio/backend/Discos/"

// var rutaBase = "/home/srmdb/Documentos/Archivos_V1S/Lab/proyecto2MIA1_vacas/backend/Discos/"

func Mkdisk(params [][]string) string {
	var size int
	var unit string = "M"
	var fit byte = 'F'

	for _, param := range params {
		key := strings.ToLower(param[1])
		val := strings.Trim(param[2], "\"")

		switch key {
		case "size":
			valInt, err := strconv.Atoi(val)
			if err != nil || valInt <= 0 {
				return "❌ Error: tamaño inválido."
			}
			size = valInt
		case "unit":
			unit = strings.ToUpper(val)
			if unit != "K" && unit != "M" {
				return "❌ Error: unidad inválida. Solo se permite K o M."
			}
		case "fit":
			val = strings.ToUpper(val)
			switch val {
			case "BF":
				fit = 'B'
			case "FF":
				fit = 'F'
			case "WF":
				fit = 'W'
			default:
				return "❌ Error: tipo de ajuste inválido. Solo BF, FF o WF."
			}

		default:
			return "⚠️ Parámetro no reconocido: " + key
		}
	}

	if size == 0 {
		return "❌ Error: el parámetro -size es obligatorio."
	}

	// Calcular tamaño total
	var totalBytes int
	if unit == "K" {
		totalBytes = size * 1024
	} else {
		totalBytes = size * 1024 * 1024
	}

	// Generar nombre de archivo (A.dsk, B.dsk, etc.)
	filename := generarNombreArchivo()
	fullPath := filepath.Join(rutaBase, filename)

	// Crear el archivo
	file, err := os.Create(fullPath)
	if err != nil {
		return "❌ Error creando el archivo: " + err.Error()
	}
	defer file.Close()

	// Llenar con ceros usando buffer de 1024 bytes
	buffer := make([]byte, 1024)
	for i := 0; i < totalBytes/1024; i++ {
		file.Write(buffer)
	}

	// Crear MBR
	mbr := structs.MBR{}
	mbr.MbrSize = int32(totalBytes)
	copy(mbr.MbrCreationDate[:], time.Now().Format("2006-01-02 15:04:05 "))
	mbr.MbrDiskSignature = rand.Int31()
	copy(mbr.DiskFit[:], string(fit))

	// Inicializar particiones vacías
	for i := 0; i < 4; i++ {
		mbr.Partitions[i] = structs.Partition{}
		copy(mbr.Partitions[i].PartStatus[:], "0")
	}

	// Escribir MBR al inicio del disco
	file.Seek(0, 0)
	err = writeStruct(file, mbr)
	if err != nil {
		return "❌ Error al escribir el MBR: " + err.Error()
	}

	return "✅ Disco creado correctamente en: " + fullPath + "\n📦 Estado actual del MBR del:\n" + mbr.PrintMBR(filename)
}

func generarNombreArchivo() string {
	files, _ := os.ReadDir(rutaBase)
	letra := 'A' + rune(len(files))
	return string(letra) + ".dsk"
}

func writeStruct(file *os.File, data interface{}) error {
	return binaryWrite(file, data)
}
