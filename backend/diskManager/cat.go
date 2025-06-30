package diskmanager

import (
	"encoding/binary"
	"fmt"
	"os"
	"proyecto1/session"
	"proyecto1/structs"
	"strings"
)

func Cat(params [][]string) {
	if !session.Sesion.LoggedIn {
		fmt.Println("❌ Error: no hay una sesión activa.")
		return
	}

	var path string
	for _, param := range params {
		if strings.ToLower(param[1]) == "file" {
			path = strings.Trim(param[2], "\"")
		}
	}
	if path == "" || path != "/users.txt" {
		fmt.Println("❌ Error: actualmente solo se puede mostrar el archivo /users.txt")
		return
	}

	id := session.Sesion.ID
	diskPath := rutaBase + string(id[0]) + ".dsk"
	file, err := os.OpenFile(diskPath, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("❌ Error al abrir el disco.")
		return
	}
	defer file.Close()

	// Leer MBR
	var mbr structs.MBR
	binary.Read(file, binary.LittleEndian, &mbr)

	// Buscar partición montada
	var part structs.Partition
	found := false
	for _, p := range mbr.Partitions {
		if string(p.PartID[:]) == id {
			part = p
			found = true
			break
		}
	}
	if !found {
		fmt.Println("❌ Error: no se encontró la partición con el ID", id)
		return
	}

	// Leer Superblock
	file.Seek(int64(part.PartStart), 0)
	var sb structs.Superblock
	binary.Read(file, binary.LittleEndian, &sb)

	// Leer inodo de users.txt (inodo 1)
	file.Seek(int64(sb.SInodeStart)+int64(1)*int64(sb.SInodeSize), 0)
	var inode structs.Inode
	binary.Read(file, binary.LittleEndian, &inode)

	// Recorrer todos los bloques apuntados por el inodo
	fmt.Println("📄 Contenido de /users.txt :")
	for i := 0; i < 15; i++ {
		/* fmt.Println(i, "****") */
		ptr := inode.IBlock[i]
		if ptr == -1 {
			continue
		}

		file.Seek(int64(sb.SBlockStart)+int64(ptr)*64, 0)
		var bloque structs.Fileblock
		binary.Read(file, binary.LittleEndian, &bloque)

		raw := string(bloque.BContent[:])
		texto := strings.Split(raw, "\x00")[0]

		lineas := strings.Split(texto, "\n")
		for _, linea := range lineas {
			linea = strings.TrimSpace(linea)
			if linea != "" {
				fmt.Println(linea)
			}
		}
	}
}
