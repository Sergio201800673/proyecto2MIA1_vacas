package diskmanager

import (
	"api-mia1/session"
	"api-mia1/structs"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

func Mkusr(params [][]string) string {
	var user, pass, grp string

	// Obtener parámetros
	for _, param := range params {
		if len(param) >= 3 {
			clave := strings.ToLower(param[1])
			valor := strings.Trim(param[2], "\"")
			switch clave {
			case "user":
				user = valor
			case "pass":
				pass = valor
			case "grp":
				grp = valor
			}
		}
	}

	if user == "" || pass == "" || grp == "" {
		return "❌ Error: Los parámetros -user, -pass y -grp son obligatorios."
	}
	if len(user) > 10 || len(pass) > 10 || len(grp) > 10 {
		return "❌ Error: Máximo 10 caracteres para usuario, contraseña y grupo."
	}

	// Solo root puede crear usuarios
	if !session.Sesion.LoggedIn || session.Sesion.Username != "root" {
		return "❌ Error: Solo el usuario root puede crear usuarios."
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

	// Verificar si el grupo existe
	grupoExiste := false
	lineaEliminada := -1
	lineas := strings.Split(contenido, "\n")
	for i, linea := range lineas {
		campos := strings.Split(linea, ",")
		if len(campos) >= 3 && strings.TrimSpace(campos[1]) == "G" && strings.TrimSpace(campos[0]) != "0" && strings.TrimSpace(campos[2]) == grp {
			grupoExiste = true
		}
		if len(campos) >= 5 && strings.TrimSpace(campos[1]) == "U" && strings.TrimSpace(campos[3]) == user {
			if strings.TrimSpace(campos[0]) == "0" {
				lineaEliminada = i // guardar índice para reutilizar
			} else {
				return "❌ Error: El usuario ya existe."
			}
		}
	}
	if !grupoExiste {
		return "❌ Error: El grupo especificado no existe."
	}

	// Buscar el siguiente UID
	maxUID := 0
	for _, linea := range lineas {
		campos := strings.Split(linea, ",")
		if len(campos) >= 5 && strings.TrimSpace(campos[1]) == "U" {
			var uid int
			fmt.Sscanf(campos[0], "%d", &uid)
			if uid > maxUID {
				maxUID = uid
			}
		}
	}
	nuevoUID := maxUID + 1

	if lineaEliminada != -1 {
		// Reutilizar línea eliminada
		campos := strings.Split(lineas[lineaEliminada], ",")
		campos[0] = fmt.Sprintf("%d", nuevoUID)
		campos[1] = "U"
		campos[2] = grp
		campos[3] = user
		campos[4] = pass
		lineas[lineaEliminada] = strings.Join(campos, ",")
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
		// Si se necesitan más bloques, asignar nuevos
		if len(bloques) > bloquesUsados {
			bitmap := make([]byte, sb.SBlocksCount)
			file.Seek(int64(sb.SBmBlockStart), 0)
			file.Read(bitmap)
			for i := 0; i < int(sb.SBlocksCount); i++ {
				if bitmap[i] == 0 {
					inodeUsers.IBlock[bloquesUsados] = int32(i)
					bitmap[i] = 1
					file.Seek(int64(sb.SBmBlockStart), 0)
					file.Write(bitmap)
					break
				}
			}
		}
		for idx, bloqueStr := range bloques {
			var bloque structs.FileBlock
			copy(bloque.Content[:], []byte(bloqueStr))
			file.Seek(int64(sb.SBlockStart)+int64(inodeUsers.IBlock[idx])*64, 0)
			binary.Write(file, binary.LittleEndian, &bloque)
		}
		inodeUsers.ISize = int32(len(nuevoContenido))
		file.Seek(int64(sb.SInodeStart)+int64(binary.Size(structs.Inode{})), 0)
		binary.Write(file, binary.LittleEndian, &inodeUsers)
		output := "✅ Usuario creado exitosamente.\nContenido actual de users.txt:\n" + nuevoContenido
		return output
	}

	// Verificar si el usuario ya existe
	for _, linea := range lineas {
		campos := strings.Split(linea, ",")
		if len(campos) >= 5 && strings.TrimSpace(campos[1]) == "U" && strings.TrimSpace(campos[0]) != "0" && strings.TrimSpace(campos[3]) == user {
			return "❌ Error: El usuario ya existe."
		}
	}

	// Buscar el siguiente UID
	maxUID = 0
	for _, linea := range lineas {
		campos := strings.Split(linea, ",")
		if len(campos) >= 5 && strings.TrimSpace(campos[1]) == "U" {
			var uid int
			fmt.Sscanf(campos[0], "%d", &uid)
			if uid > maxUID {
				maxUID = uid
			}
		}
	}
	nuevoUID = maxUID + 1

	// Agregar el usuario al contenido
	nuevaLinea := fmt.Sprintf("%d,U,%s,%s,%s\n", nuevoUID, grp, user, pass)
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

	output := "✅ Usuario creado exitosamente.\nContenido actual de users.txt:\n" + nuevoContenido
	return output
}
