package diskmanager

import (
	"encoding/binary"
	"fmt"
	"os"
	"proyecto1/structs"
	"strconv"
	"strings"
)

func Fdisk(params [][]string) {
	var size int
	var unit string = "K"
	var fit byte = 'W'
	var ptype byte = 'P'
	var name string = ""
	var driveletter string = ""

	var deleteFlag bool = false

	var addValue int
	var hasAdd bool = false

	// Leer parámetros
	for _, param := range params {
		key := strings.ToLower(param[1])
		val := strings.Trim(param[2], "\"")

		switch key {
		case "size":
			valInt, err := strconv.Atoi(val)
			if err != nil || valInt <= 0 {
				fmt.Println("❌ Error: tamaño de partición inválido.")
				return
			}
			size = valInt
		case "unit":
			val = strings.ToUpper(val)
			if val != "B" && val != "K" && val != "M" {
				fmt.Println("❌ Error: unidad inválida. Solo B, K o M.")
				return
			}
			unit = val
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
				fmt.Println("❌ Error: ajuste inválido. Solo BF, FF o WF.")
				return
			}
		case "type":
			val = strings.ToUpper(val)
			if val != "P" && val != "E" {
				fmt.Println("❌ Error: tipo de partición inválido. Solo P o E.")
				return
			}
			ptype = val[0]
		case "name":
			name = val
		case "driveletter":
			driveletter = strings.ToUpper(val)
		case "delete":
			if strings.ToLower(val) == "full" {
				deleteFlag = true
			} else {
				fmt.Println("❌ Error: valor inválido para -delete. Solo se permite 'full'.")
				return
			}
		case "add":
			amount, err := strconv.Atoi(val)
			if err != nil {
				fmt.Println("❌ Error: el valor de -add debe ser numérico.")
				return
			}
			addValue = amount
			hasAdd = true
		default:
			fmt.Println("⚠️ Parámetro no reconocido:", key)
		}
	}

	if !deleteFlag && !hasAdd && (size <= 0 || name == "" || driveletter == "") {
		fmt.Println("❌ Error: parámetros obligatorios faltantes (-size, -name, -driveletter).")
		return
	}

	if deleteFlag {
		if name == "" || driveletter == "" {
			fmt.Println("❌ Error: -delete requiere -name y -driveletter.")
			return
		}

		// Abrir disco
		path := rutaBase + driveletter + ".dsk"
		file, err := os.OpenFile(path, os.O_RDWR, 0666)
		if err != nil {
			fmt.Println("❌ Error: no se encontró el disco", driveletter+".dsk")
			return
		}
		defer file.Close()

		var mbr structs.MBR
		err = binary.Read(file, binary.LittleEndian, &mbr)
		if err != nil {
			fmt.Println("❌ Error al leer MBR.")
			return
		}

		// Buscar partición por nombre
		found := false
		for i := 0; i < 4; i++ {
			if getCleanName(mbr.Partitions[i].PartName) == name && mbr.Partitions[i].PartStatus[0] == '1' {
				found = true
				// Borrar bytes
				file.Seek(int64(mbr.Partitions[i].PartStart), 0)
				cero := make([]byte, mbr.Partitions[i].PartSize)
				file.Write(cero)

				// Marcar como eliminada
				mbr.Partitions[i] = structs.Partition{} // limpia todo
				break
			}
		}

		if !found {
			fmt.Println("❌ Error: no se encontró una partición activa con el nombre", name)
			return
		}

		// Guardar nuevo MBR
		file.Seek(0, 0)
		binary.Write(file, binary.LittleEndian, &mbr)
		fmt.Println("🗑️ Partición", name, "eliminada correctamente.")
		mbr.PrintMBR()
		return
	}

	if hasAdd {
		if name == "" || driveletter == "" {
			fmt.Println("❌ Error: -add requiere -name y -driveletter.")
			return
		}

		// Unidad
		addBytes := int32(addValue)
		switch unit {
		case "K":
			addBytes *= 1024
		case "M":
			addBytes *= 1024 * 1024
		}

		path := rutaBase + driveletter + ".dsk"
		file, err := os.OpenFile(path, os.O_RDWR, 0666)
		if err != nil {
			fmt.Println("❌ Error: disco no encontrado.")
			return
		}
		defer file.Close()

		var mbr structs.MBR
		err = binary.Read(file, binary.LittleEndian, &mbr)
		if err != nil {
			fmt.Println("❌ Error al leer MBR.")
			return
		}

		found := false
		for i := 0; i < 4; i++ {
			part := &mbr.Partitions[i]
			if getCleanName(part.PartName) == name && part.PartStatus[0] == '1' {
				found = true
				if addBytes > 0 {
					// Verificar que hay espacio libre después
					end := part.PartStart + part.PartSize
					minStart := mbr.MbrSize
					for j := 0; j < 4; j++ {
						if mbr.Partitions[j].PartStatus[0] == '1' && mbr.Partitions[j].PartStart > end {
							if mbr.Partitions[j].PartStart < minStart {
								minStart = mbr.Partitions[j].PartStart
							}
						}
					}
					if end+addBytes > minStart {
						fmt.Println("❌ Error: no hay espacio suficiente para ampliar la partición.")
						return
					}
					part.PartSize += addBytes
				} else {
					// Verificar que no quede tamaño negativo
					if part.PartSize+addBytes <= 0 {
						fmt.Println("❌ Error: no se puede reducir tanto la partición.")
						return
					}
					part.PartSize += addBytes
				}
				break
			}
		}

		if !found {
			fmt.Println("❌ Error: no se encontró la partición", name)
			return
		}

		file.Seek(0, 0)
		binary.Write(file, binary.LittleEndian, &mbr)
		fmt.Println("✅ Tamaño de la partición", name, "modificado correctamente.")
		mbr.PrintMBR()
		return
	}

	// Calcular tamaño real
	var realSize int32 = int32(size)
	switch unit {
	case "B":
		// Nada
	case "K":
		realSize *= 1024
	case "M":
		realSize *= 1024 * 1024
	}

	// Leer el disco
	path := rutaBase + driveletter + ".dsk"
	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("❌ Error: el disco", driveletter+".dsk", "no existe.")
		return
	}
	defer file.Close()

	// Leer MBR
	var mbr structs.MBR
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		fmt.Println("❌ Error al leer el MBR:", err)
		return
	}

	// Verificar que no exista partición con el mismo nombre
	for _, p := range mbr.Partitions {
		if getCleanName(p.PartName) == name {
			fmt.Println("❌ Error: ya existe una partición con el nombre", name)
			return
		}
	}

	// Verificar si hay espacio disponible
	var start int32 = int32(binary.Size(mbr))
	for i := 0; i < 4; i++ {
		if mbr.Partitions[i].PartStatus[0] == '1' {
			end := mbr.Partitions[i].PartStart + mbr.Partitions[i].PartSize
			if end > start {
				start = end
			}
		}
	}

	if start+realSize > mbr.MbrSize {
		fmt.Println("❌ Error: no hay suficiente espacio libre para la partición.")
		return
	}

	// Verificar si ya hay partición extendida
	if ptype == 'E' {
		for _, p := range mbr.Partitions {
			if p.PartType[0] == 'E' {
				fmt.Println("❌ Error: ya existe una partición extendida.")
				return
			}
		}
	}

	// Buscar un espacio libre en particiones
	for i := 0; i < 4; i++ {
		if mbr.Partitions[i].PartStatus[0] != '1' {
			// Llenar la partición
			copy(mbr.Partitions[i].PartStatus[:], "1")
			mbr.Partitions[i].PartStart = start
			mbr.Partitions[i].PartSize = realSize
			mbr.Partitions[i].PartType[0] = ptype
			mbr.Partitions[i].PartFit[0] = fit
			copy(mbr.Partitions[i].PartName[:], name)
			mbr.Partitions[i].PartCorrelative = int32(i + 1)
			break
		}
		if i == 3 {
			fmt.Println("❌ Error: ya existen 4 particiones.")
			return
		}
	}

	// Reescribir MBR
	file.Seek(0, 0)
	err = binary.Write(file, binary.LittleEndian, &mbr)
	if err != nil {
		fmt.Println("❌ Error al escribir el MBR:", err)
		return
	}

	fmt.Println("📦 Estado actual del MBR:")
	mbr.PrintMBR()

	fmt.Println("✅ Partición", name, "creada correctamente en el disco", driveletter+".dsk")
}

func getCleanName(nameBytes [16]byte) string {
	return strings.Split(string(nameBytes[:]), "\x00")[0]
}
