package diskmanager

import (
	"api-mia1/structs"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

func Mount(params [][]string) string {
	var name string
	var driveletter string

	// Leer parámetros
	for _, param := range params {
		key := strings.ToLower(param[1])
		val := strings.Trim(param[2], "\"")

		switch key {
		case "name":
			name = val
		case "driveletter":
			driveletter = strings.ToUpper(val)
		default:
			return "⚠️ Parámetro no reconocido:" + key
		}
	}

	if name == "" || driveletter == "" {
		return "❌ Error: -name y -driveletter son obligatorios."
	}

	path := rutaBase + driveletter + ".dsk"
	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return "❌ Error: no se encontró el disco " + driveletter + ".dsk"
	}
	defer file.Close()

	var mbr structs.MBR
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		return "❌ Error al leer el MBR del disco."
	}

	output := ""

	found := false
	for i := 0; i < 4; i++ {
		part := &mbr.Partitions[i]
		if getCleanName(part.PartName) == name && part.PartStatus[0] == '1' {
			if part.PartType[0] != 'P' {
				return "❌ Solo se pueden montar particiones primarias."
			}
			part.PartStatus[0] = '2' // montada (diferente de '1')
			part.PartCorrelative = int32(i + 1)

			// Guardar en disco
			file.Seek(0, 0)
			binary.Write(file, binary.LittleEndian, &mbr)

			// Generar ID
			correlativo := i + 1
			id := fmt.Sprintf("%s%d%s", driveletter, correlativo, carnet[len(carnet)-2:])
			copy(part.PartID[:], id[:4])
			file.Seek(0, 0)
			binary.Write(file, binary.LittleEndian, &mbr)

			// Guardar en memoria
			m := Montaje{
				ID:              id,
				DriveLetter:     driveletter,
				PartitionName:   name,
				PartCorrelative: correlativo,
			}
			Montajes = append(Montajes, m)

			output = "✅ Partición montada exitosamente.\n🆔 ID asignado: " + id

			found = true
			break
		}
	}
	if !found {
		return "❌ Error: no se encontró la partición con ese nombre en el disco."
	}
	return output + "\n📦 Estado actual del MBR:\n" + mbr.PrintMBR(driveletter)
}
