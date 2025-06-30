package diskmanager

import (
	"encoding/binary"
	"fmt"
	"os"
	"proyecto1/structs"
	"strings"
)

func Unmount(params [][]string) {
	var id string

	for _, param := range params {
		if strings.ToLower(param[1]) == "id" {
			id = strings.ToUpper(strings.Trim(param[2], "\""))
		}
	}

	if id == "" {
		fmt.Println("❌ Error: parámetro -id es obligatorio.")
		return
	}

	driveletter := string(id[0])
	path := rutaBase + driveletter + ".dsk"

	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("❌ Error: no se pudo abrir el disco", path)
		return
	}
	defer file.Close()

	var mbr structs.MBR
	if err := binary.Read(file, binary.LittleEndian, &mbr); err != nil {
		fmt.Println("❌ Error al leer el MBR.")
		return
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
		fmt.Println("❌ Error: no se encontró ninguna partición con el ID", id)
		return
	}

	file.Seek(0, 0)
	binary.Write(file, binary.LittleEndian, &mbr)

	fmt.Println("✅ Partición", id, "desmontada exitosamente.")
	mbr.PrintMBR()
}
