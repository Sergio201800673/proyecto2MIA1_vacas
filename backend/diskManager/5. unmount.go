package diskmanager

import (
	"api-mia1/structs"
	"encoding/binary"
	"os"
	"strings"
)

func Unmount(params [][]string) string {
	var id string

	for _, param := range params {
		if strings.ToLower(param[1]) == "id" {
			id = strings.ToUpper(strings.Trim(param[2], "\""))
		}
	}

	if id == "" {
		return "❌ Error: parámetro -id es obligatorio."
	}

	driveletter := string(id[0])
	path := rutaBase + driveletter + ".dsk"

	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return "❌ Error: no se pudo abrir el disco " + path
	}
	defer file.Close()

	var mbr structs.MBR
	if err := binary.Read(file, binary.LittleEndian, &mbr); err != nil {
		return "❌ Error al leer el MBR."
	}

	found := false
	for i := 0; i < 4; i++ {
		part := &mbr.Partitions[i]
		if string(part.PartID[:]) == id {
			part.PartStatus[0] = '1' // activa pero no montada
			part.PartCorrelative = 0
			copy(part.PartID[:], "")
			found = true
			break
		}
	}

	if !found {
		return "❌ Error: no se encontró ninguna partición con el ID " + id
	}

	file.Seek(0, 0)
	binary.Write(file, binary.LittleEndian, &mbr)

	return "✅ Partición " + id + " desmontada exitosamente.\n📦 Estado actual del MBR:\n" + mbr.PrintMBR(driveletter)
}
