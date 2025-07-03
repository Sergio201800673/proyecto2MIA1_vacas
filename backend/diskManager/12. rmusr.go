package diskmanager

import (
	"api-mia1/session"
	"api-mia1/structs"
	"encoding/binary"
	"os"
	"strings"
)

func Rmusr(params [][]string) string {
	var user string

	// Obtener el nombre del usuario
	for _, param := range params {
		if len(param) >= 3 && strings.ToLower(param[1]) == "user" {
			user = strings.Trim(param[2], "\"")
		}
	}

	if user == "" {
		return "❌ Error: El parámetro -user es obligatorio."
	}

	// Solo root puede eliminar usuarios
	if !session.Sesion.LoggedIn || session.Sesion.Username != "root" {
		return "❌ Error: Solo el usuario root puede eliminar usuarios."
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
	for _, b := range inodeUsers.IBlock {
		if b == -1 {
			continue
		}
		var bloque structs.FileBlock
		file.Seek(int64(sb.SBlockStart)+int64(b)*64, 0)
		binary.Read(file, binary.LittleEndian, &bloque)
		contenido += string(bloque.Content[:])
	}

	// Buscar y marcar el usuario como eliminado (UID=0)
	lineas := strings.Split(contenido, "\n")
	usuarioEncontrado := false
	for i, linea := range lineas {
		campos := strings.Split(linea, ",")
		if len(campos) >= 5 && strings.TrimSpace(campos[1]) == "U" && strings.TrimSpace(campos[0]) != "0" && strings.TrimSpace(campos[3]) == user {
			campos[0] = "0" // Marcar como eliminado
			lineas[i] = strings.Join(campos, ",")
			usuarioEncontrado = true
			break
		}
	}
	if !usuarioEncontrado {
		return "❌ Error: El usuario no existe."
	}

	nuevoContenido := strings.Join(lineas, "\n")

	// Repartir el contenido en bloques de 64 bytes
	bloques := []string{}
	for i := 0; i < len(nuevoContenido); i += 64 {
		fin := i + 64
		if fin > len(nuevoContenido) {
			fin = len(nuevoContenido)
		}
		bloques = append(bloques, nuevoContenido[i:fin])
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

	output := "✅ Usuario eliminado exitosamente.\nContenido actual de users.txt:\n" + nuevoContenido
	return output
}
