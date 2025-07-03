package diskmanager

import (
	"api-mia1/session"
	"api-mia1/structs"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

func Mkgrp(params [][]string) string {
	var name string

	// Obtener el nombre del grupo
	for _, param := range params {
		if len(param) >= 3 && strings.ToLower(param[1]) == "name" {
			name = strings.Trim(param[2], "\"")
		}
	}

	if name == "" {
		return "❌ Error: El parámetro -name es obligatorio."
	}

	// Solo root puede crear grupos
	if !session.Sesion.LoggedIn || session.Sesion.Username != "root" {
		return "❌ Error: Solo el usuario root puede crear grupos."
	}

	// Abrir disco
	drive := string(session.Sesion.ID[0])
	path := rutaBase + drive + ".dsk"
	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return "❌ Error: disco no encontrado."
	}
	defer file.Close()

	// Leer MBR
	var mbr structs.MBR
	binary.Read(file, binary.LittleEndian, &mbr)

	// Buscar partición por ID
	var part structs.Partition
	found := false
	for _, p := range mbr.Partitions {
		if string(p.PartID[:]) == session.Sesion.ID {
			part = p
			found = true
			break
		}
	}
	if !found {
		return "❌ Error: partición no encontrada."
	}

	// Leer Superblock
	sb := structs.Superblock{}
	file.Seek(int64(part.PartStart), 0)
	binary.Read(file, binary.LittleEndian, &sb)

	// Leer inodo de users.txt (usualmente inodo 1)
	inodeUsers := structs.Inode{}
	file.Seek(int64(sb.SInodeStart)+int64(binary.Size(structs.Inode{})), 0)
	binary.Read(file, binary.LittleEndian, &inodeUsers)

	// Leer todos los bloques asignados a users.txt
	contenido := ""
	bloquesUsados := 0
	for _, b := range inodeUsers.IBlock {
		if b == -1 {
			continue
		}
		bloquesUsados++
		var bloque structs.FileBlock
		file.Seek(int64(sb.SBlockStart)+int64(b)*64, 0)
		binary.Read(file, binary.LittleEndian, &bloque)
		contenido += string(bloque.Content[:])
	}

	fmt.Println("Contenido actual de users.txt:")
	fmt.Println(contenido)

	// Verificar si el grupo ya existe
	lineas := strings.Split(contenido, "\n")
	for _, linea := range lineas {
		campos := strings.Split(linea, ",")
		if len(campos) >= 3 && strings.TrimSpace(campos[1]) == "G" && strings.TrimSpace(campos[2]) == name {
			return "❌ Error: El grupo ya existe."
		}
	}

	// Buscar el siguiente GID
	maxGID := 0
	for _, linea := range lineas {
		campos := strings.Split(linea, ",")
		if len(campos) >= 3 && strings.TrimSpace(campos[1]) == "G" {
			var gid int
			fmt.Sscanf(campos[0], "%d", &gid)
			if gid > maxGID {
				maxGID = gid
			}
		}
	}
	nuevoGID := maxGID + 1

	// Agregar el grupo al contenido
	nuevaLinea := fmt.Sprintf("%d,G,%s\n", nuevoGID, name)
	nuevoContenido := contenido + nuevaLinea

	// Repartir el contenido en bloques de 64 bytes
	bloques := []string{}
	for i := 0; i < len(nuevoContenido); i += 64 {
		fin := i + 64
		if fin > len(nuevoContenido) {
			fin = len(nuevoContenido)
		}
		bloques = append(bloques, nuevoContenido[i:fin])
	}

	// Si se necesitan más bloques, asignar nuevos
	if len(bloques) > bloquesUsados {
		// Buscar un bloque libre en el bitmap de bloques
		bitmap := make([]byte, sb.SBlocksCount)
		file.Seek(int64(sb.SBmBlockStart), 0)
		file.Read(bitmap)
		for i := 0; i < int(sb.SBlocksCount); i++ {
			if bitmap[i] == 0 {
				// Asignar este bloque
				inodeUsers.IBlock[bloquesUsados] = int32(i)
				bitmap[i] = 1
				file.Seek(int64(sb.SBmBlockStart), 0)
				file.Write(bitmap)
				break
			}
		}
	}

	// Escribir los bloques actualizados
	for idx, bloqueStr := range bloques {
		var bloque structs.FileBlock
		copy(bloque.Content[:], []byte(bloqueStr))
		file.Seek(int64(sb.SBlockStart)+int64(inodeUsers.IBlock[idx])*64, 0)
		binary.Write(file, binary.LittleEndian, &bloque)
	}

	// Actualizar el tamaño del archivo en el inodo
	inodeUsers.ISize = int32(len(nuevoContenido))
	file.Seek(int64(sb.SInodeStart)+int64(binary.Size(structs.Inode{})), 0)
	binary.Write(file, binary.LittleEndian, &inodeUsers)

	output := "✅ Grupo creado exitosamente. \nContenido actual de users.txt: \n" + nuevoContenido

	return output
}
