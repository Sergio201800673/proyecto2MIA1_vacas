package diskmanager

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"proyecto1/session"
	"proyecto1/structs"
)

func Rmusr(params [][]string) {
	if !session.Sesion.LoggedIn {
		fmt.Println("❌ Error: no hay sesión activa.")
		return
	}
	if session.Sesion.Username != "root" {
		fmt.Println("❌ Error: solo el usuario root puede eliminar usuarios.")
		return
	}

	var username string
	for _, param := range params {
		if strings.ToLower(param[1]) == "user" {
			username = strings.Trim(param[2], "\"")
		}
	}
	if username == "" {
		fmt.Println("❌ Error: el parámetro -user es obligatorio.")
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

	// Leer inodo de users.txt (inodo 1)
	var inode structs.Inode
	file.Seek(int64(sb.SInodeStart)+int64(1)*int64(sb.SInodeSize), 0)
	binary.Read(file, binary.LittleEndian, &inode)

	// Leer bloques y concatenar contenido
	var contenido string
	var bloquesLeidos [15]structs.Fileblock
	for i := 0; i < 15; i++ {
		ptr := inode.IBlock[i]
		if ptr == -1 {
			continue
		}
		file.Seek(int64(sb.SBlockStart)+int64(ptr)*64, 0)
		binary.Read(file, binary.LittleEndian, &bloquesLeidos[i])
		contenido += string(bloquesLeidos[i].BContent[:])
	}
	contenido = strings.TrimRight(contenido, "\x00")
	lineas := strings.Split(contenido, "\n")

	usuarioEncontrado := false
	var nuevoContenido string

	for _, linea := range lineas {
		original := linea
		linea = strings.TrimSpace(linea)
		if linea == "" {
			continue
		}

		datos := strings.Split(linea, ",")
		if len(datos) >= 5 && strings.TrimSpace(datos[1]) == "U" {
			usr := strings.TrimSpace(datos[3])
			if usr == username {
				datos[0] = "0"
				usuarioEncontrado = true
				nuevoContenido += strings.Join(datos, ",") + "\n"
				continue
			}
		}
		nuevoContenido += original + "\n"
	}

	if !usuarioEncontrado {
		fmt.Println("❌ Error: el usuario", username, "no existe o ya está eliminado.")
		return
	}

	// Repartir contenido actualizado en bloques (máx. 64 bytes cada uno)
	nuevasLineas := strings.Split(nuevoContenido, "\n")
	bloqueActual := ""
	bloqueIndex := 0

	for _, linea := range nuevasLineas {
		if linea == "" {
			continue
		}
		if len(bloqueActual)+len(linea)+1 > 64 {
			copy(bloquesLeidos[bloqueIndex].BContent[:], []byte(bloqueActual))
			file.Seek(int64(sb.SBlockStart)+int64(inode.IBlock[bloqueIndex])*64, 0)
			binary.Write(file, binary.LittleEndian, &bloquesLeidos[bloqueIndex])
			bloqueIndex++
			bloqueActual = ""
		}
		bloqueActual += linea + "\n"
	}
	if bloqueActual != "" && bloqueIndex < 15 && inode.IBlock[bloqueIndex] != -1 {
		copy(bloquesLeidos[bloqueIndex].BContent[:], []byte(bloqueActual))
		file.Seek(int64(sb.SBlockStart)+int64(inode.IBlock[bloqueIndex])*64, 0)
		binary.Write(file, binary.LittleEndian, &bloquesLeidos[bloqueIndex])
	}

	fmt.Println("🗑️ Usuario", username, "eliminado (marcado como inactivo).")
}
