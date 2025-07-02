package diskmanager

import (
	"api-mia1/session"
	"api-mia1/structs"
	"encoding/binary"
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

	// Leer el archivo users.txt
	// Se encuentra en el primer bloque de datos (por simplicidad)
	var contenido [64]byte
	sb := structs.Superblock{}
	file.Seek(int64(part.PartStart), 0)
	binary.Read(file, binary.LittleEndian, &sb)

	// Primer bloque con users.txt
	file.Seek(int64(sb.SBlockStart), 0)
	binary.Read(file, binary.LittleEndian, &contenido)

	texto := string(contenido[:])
	lineas := strings.Split(texto, "\n")

	for _, linea := range lineas {
		if strings.HasPrefix(linea, "1,U,") {
			datos := strings.Split(linea, ",")
			if len(datos) >= 5 {
				userName := datos[2]
				userGroup := datos[1]
				userPass := datos[4]
				if user == userName && pass == userPass {
					session.Sesion = session.UsuarioActivo{
						Username: user,
						Group:    userGroup,
						ID:       id,
						LoggedIn: true,
					}
					return "✅ Sesión iniciada exitosamente como " + user + " en " + id
				}
			}
		}
	}
	return "❌ Usuario o contraseña incorrectos."
}
