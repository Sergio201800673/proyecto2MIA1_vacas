package diskmanager

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"proyecto1/session"
	"proyecto1/structs"
)

func Mkgrp(params [][]string) {
	if !session.Sesion.LoggedIn {
		fmt.Println("❌ Error: no hay una sesión activa.")
		return
	}
	if session.Sesion.Username != "root" {
		fmt.Println("❌ Error: solo el usuario root puede crear grupos.")
		return
	}

	var name string
	for _, param := range params {
		if strings.ToLower(param[1]) == "name" {
			name = strings.Trim(param[2], "\"")
		}
	}
	if name == "" {
		fmt.Println("❌ Error: el parámetro -name es obligatorio.")
		return
	}

	id := session.Sesion.ID
	diskPath := rutaBase + string(id[0]) + ".dsk"

	file, err := os.OpenFile(diskPath, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("❌ Error al abrir disco.")
		return
	}
	defer file.Close()

	var mbr structs.MBR
	binary.Read(file, binary.LittleEndian, &mbr)

	var part structs.Partition
	for _, p := range mbr.Partitions {
		if string(p.PartID[:]) == id {
			part = p
			break
		}
	}

	var sb structs.Superblock
	file.Seek(int64(part.PartStart), 0)
	binary.Read(file, binary.LittleEndian, &sb)

	// Leer inodo de users.txt
	var inode structs.Inode
	file.Seek(int64(sb.SInodeStart)+int64(1)*int64(sb.SInodeSize), 0)
	binary.Read(file, binary.LittleEndian, &inode)

	// Leer contenido actual de users.txt
	var contenido string
	for i := 0; i < 15; i++ {
		ptr := inode.IBlock[i]
		if ptr == -1 {
			continue
		}
		file.Seek(int64(sb.SBlockStart)+int64(ptr)*64, 0)
		var bloque structs.Fileblock
		binary.Read(file, binary.LittleEndian, &bloque)
		contenido += string(bloque.BContent[:])
	}
	contenido = strings.TrimRight(contenido, "\x00")
	lineas := strings.Split(contenido, "\n")

	// Validar grupo duplicado
	for _, linea := range lineas {
		linea = strings.TrimSpace(linea)
		datos := strings.Split(linea, ",")
		if len(datos) >= 3 && strings.TrimSpace(datos[1]) == "G" {
			if strings.EqualFold(strings.TrimSpace(datos[2]), name) {
				fmt.Println("❌ Error: el grupo", name, "ya existe.")
				return
			}
		}
	}

	// Obtener nuevo ID
	maxID := 0
	for _, linea := range lineas {
		linea = strings.TrimSpace(linea)
		datos := strings.Split(linea, ",")
		if len(datos) >= 1 {
			var tempID int
			fmt.Sscanf(strings.TrimSpace(datos[0]), "%d", &tempID)
			if tempID > maxID {
				maxID = tempID
			}
		}
	}
	nuevoID := maxID + 1
	nuevaLinea := fmt.Sprintf("%d,G,%s\n", nuevoID, name)

	// Escribir nueva línea en bloque disponible
	ok, inode, sb := escribirEnBloqueDisponible(file, sb, inode, nuevaLinea, part.PartStart)
	if !ok {
		fmt.Println("❌ Error: no hay bloques disponibles para escribir.")
		return
	}

	// Guardar cambios en inodo y superblock
	file.Seek(int64(sb.SInodeStart)+int64(1)*int64(sb.SInodeSize), 0)
	binary.Write(file, binary.LittleEndian, &inode)

	file.Seek(int64(part.PartStart), 0)
	binary.Write(file, binary.LittleEndian, &sb)

	fmt.Println("✅ Grupo", name, "creado exitosamente.")
}

func escribirEnBloqueDisponible(file *os.File, sb structs.Superblock, inode structs.Inode, nuevaLinea string, partStart int32) (bool, structs.Inode, structs.Superblock) {
	var bloque structs.Fileblock
	for i := 0; i < 15; i++ {
		ptr := inode.IBlock[i]
		if ptr == -1 {
			// Nuevo bloque
			ptr = sb.SFirstBlock
			inode.IBlock[i] = ptr
			sb.SFirstBlock += 1
			sb.SFreeBlocksCount -= 1

			bitmapPos := sb.SBmBlockStart + ptr
			file.Seek(int64(bitmapPos), 0)
			file.Write([]byte{1})

			// Escribir directamente la nueva línea
			copy(bloque.BContent[:], []byte(nuevaLinea))
			file.Seek(int64(sb.SBlockStart)+int64(ptr)*64, 0)
			binary.Write(file, binary.LittleEndian, &bloque)

			inode.ISize += int32(len(nuevaLinea))

			return true, inode, sb
		} else {
			// Leer y verificar si hay espacio
			file.Seek(int64(sb.SBlockStart)+int64(ptr)*64, 0)
			binary.Read(file, binary.LittleEndian, &bloque)
			texto := string(bloque.BContent[:])
			texto = strings.TrimRight(texto, "\x00")
			if len(texto)+len(nuevaLinea) <= 64 {
				nuevoTexto := texto + nuevaLinea

				for i := range bloque.BContent {
					bloque.BContent[i] = 0
				}

				copy(bloque.BContent[:], []byte(nuevoTexto))
				file.Seek(int64(sb.SBlockStart)+int64(ptr)*64, 0)
				binary.Write(file, binary.LittleEndian, &bloque)

				inode.ISize += int32(len(nuevaLinea))

				return true, inode, sb
			}
		}
	}
	return false, inode, sb
}
