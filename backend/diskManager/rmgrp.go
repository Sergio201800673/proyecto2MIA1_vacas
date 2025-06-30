package diskmanager

import (
	"encoding/binary"
	"fmt"
	"os"
	"proyecto1/session"
	"proyecto1/structs"
	"strings"
)

func Rmgrp(params [][]string) {
	if !session.Sesion.LoggedIn {
		fmt.Println("❌ Error: no hay sesión activa.")
		return
	}
	if session.Sesion.Username != "root" {
		fmt.Println("❌ Error: solo el usuario root puede eliminar grupos.")
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
		fmt.Println("❌ Error al abrir el disco.")
		return
	}
	defer file.Close()

	// Leer MBR
	var mbr structs.MBR
	binary.Read(file, binary.LittleEndian, &mbr)

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
		fmt.Println("❌ Partición no encontrada.")
		return
	}

	// Leer Superblock
	var sb structs.Superblock
	file.Seek(int64(part.PartStart), 0)
	binary.Read(file, binary.LittleEndian, &sb)

	// Leer el bloque de users.txt
	file.Seek(int64(sb.SBlockStart), 0)
	var bloque structs.Fileblock
	binary.Read(file, binary.LittleEndian, &bloque)

	contenido := string(bloque.BContent[:])
	contenido = strings.TrimRight(contenido, "\x00")
	lineas := strings.Split(contenido, "\n")

	grupoEncontrado := false
	var nuevoContenido string

	for _, linea := range lineas {
		original := linea
		linea = strings.TrimSpace(linea)
		if linea == "" {
			continue
		}

		datos := strings.Split(linea, ",")
		if len(datos) >= 3 && strings.TrimSpace(datos[1]) == "G" {
			grupo := strings.TrimSpace(datos[2])
			if grupo == name {
				// Marcar como eliminado
				datos[0] = "0"
				grupoEncontrado = true
				nuevoContenido += strings.Join(datos, ",") + "\n"
				continue
			}
		}

		nuevoContenido += original + "\n"
	}

	if !grupoEncontrado {
		fmt.Println("❌ Error: el grupo", name, "no existe o ya está eliminado.")
		return
	}

	if len(nuevoContenido) > 64 {
		fmt.Println("❌ Error: el archivo users.txt está lleno. No se puede actualizar.")
		return
	}

	// Escribir nuevo contenido
	var nuevoBloque structs.Fileblock
	copy(nuevoBloque.BContent[:], []byte(nuevoContenido))
	file.Seek(int64(sb.SBlockStart), 0)
	binary.Write(file, binary.LittleEndian, &nuevoBloque)

	fmt.Println("🗑️ Grupo", name, "eliminado (marcado como inactivo).")
}
