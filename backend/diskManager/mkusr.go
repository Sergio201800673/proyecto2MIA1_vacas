package diskmanager

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"proyecto1/session"
	"proyecto1/structs"
)

func Mkusr(params [][]string) {
	if !session.Sesion.LoggedIn {
		fmt.Println("❌ Error: no hay sesión activa.")
		return
	}
	if session.Sesion.Username != "root" {
		fmt.Println("❌ Error: solo el usuario root puede crear usuarios.")
		return
	}

	var user, pass, grp string
	for _, param := range params {
		switch strings.ToLower(param[1]) {
		case "user":
			user = strings.Trim(param[2], "\"")
		case "pass":
			pass = strings.Trim(param[2], "\"")
		case "grp":
			grp = strings.Trim(param[2], "\"")
		}
	}
	if user == "" || pass == "" || grp == "" {
		fmt.Println("❌ Error: parámetros -user, -pass y -grp son obligatorios.")
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

	// Leer inodo de users.txt (inodo 1)
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

	// Validar que el grupo existe y está activo
	grupoExiste := false
	for _, linea := range lineas {
		linea = strings.TrimSpace(linea)
		datos := strings.Split(linea, ",")
		if len(datos) >= 3 && strings.TrimSpace(datos[1]) == "G" {
			if strings.EqualFold(strings.TrimSpace(datos[2]), grp) {
				grupoExiste = true
				break
			}
		}
	}
	if !grupoExiste {
		fmt.Println("❌ Error: el grupo", grp, "no existe o está inactivo.")
		return
	}

	// Validar que el usuario no exista
	for _, linea := range lineas {
		linea = strings.TrimSpace(linea)
		datos := strings.Split(linea, ",")
		if len(datos) >= 5 && strings.TrimSpace(datos[1]) == "U" {
			if strings.EqualFold(strings.TrimSpace(datos[3]), user) {
				fmt.Println("❌ Error: el usuario", user, "ya existe.")
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
	nuevaLinea := fmt.Sprintf("%d,U,%s,%s,%s\n", nuevoID, grp, user, pass)

	// Escribir la nueva línea
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

	fmt.Println("✅ Usuario", user, "creado exitosamente.")
}
