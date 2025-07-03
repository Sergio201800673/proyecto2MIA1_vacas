package diskmanager

import (
	"api-mia1/session"
	"api-mia1/structs"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

func Login(params [][]string) string {
	var user, pass, id string

	// Leer parámetros
	for _, param := range params {
		switch strings.ToLower(param[1]) {
		case "user":
			user = strings.Trim(param[2], "\"")
		case "pass":
			pass = strings.Trim(param[2], "\"")
		case "id":
			id = strings.Trim(param[2], "\"")
		}
	}

	// Validación básica
	if user == "" || pass == "" || id == "" {
		return "❌ Error: los parámetros -user, -pass e -id son obligatorios."
	}

	// Revisar si ya hay alguien logueado
	if session.Sesion.LoggedIn {
		return "❌ Ya hay una sesión activa. Debe hacer logout primero."
	}

	// Abrir disco
	drive := string(id[0])
	path := rutaBase + drive + ".dsk"

	file, err := os.Open(path)
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
		if string(p.PartID[:]) == id {
			part = p
			found = true
			break
		}
	}
	if !found {
		return "❌ Error: partición no encontrada."
	}

	// Leer el superbloque
	sb := structs.Superblock{}
	file.Seek(int64(part.PartStart), 0)
	binary.Read(file, binary.LittleEndian, &sb)

	// Leer inodo de users.txt (usualmente inodo 1)
	inodeUsers := structs.Inode{}
	file.Seek(int64(sb.SInodeStart)+int64(binary.Size(structs.Inode{})), 0)
	binary.Read(file, binary.LittleEndian, &inodeUsers)

	// Leer todos los bloques asignados a users.txt
	texto := ""
	for _, b := range inodeUsers.IBlock {
		if b == -1 {
			continue
		}
		var bloque structs.FileBlock
		file.Seek(int64(sb.SBlockStart)+int64(b)*64, 0)
		binary.Read(file, binary.LittleEndian, &bloque)
		texto += string(bloque.Content[:])
	}
	lineas := strings.Split(texto, "\n")

	// Buscar usuario en users.txt
	usuarioEncontrado := false

	fmt.Println(lineas, "Archivo users.txt impreso desde login")

	for _, linea := range lineas {
		campos := strings.Split(linea, ",")
		if len(campos) == 5 && strings.TrimSpace(campos[1]) == "U" {
			if strings.TrimSpace(campos[0]) == "0" {
				continue // usuario eliminado, ignorar
			}
			userName := strings.TrimSpace(campos[3])
			userPass := strings.TrimSpace(campos[4])
			userGroup := strings.TrimSpace(campos[2])
			if user == userName {
				usuarioEncontrado = true
				if pass == userPass {
					session.Sesion = session.UsuarioActivo{
						Username: user,
						Group:    userGroup,
						ID:       id,
						LoggedIn: true,
					}
					return "✅ Sesión iniciada exitosamente como " + user + " en " + id
				} else {
					return "❌ Contraseña incorrecta."
				}
			}
		}
	}
	if usuarioEncontrado {
		return "❌ Contraseña incorrecta."
	}
	return "❌ Usuario no existe."
}
