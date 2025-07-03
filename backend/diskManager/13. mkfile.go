package diskmanager

import (
	"api-mia1/session"
	"api-mia1/structs"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func Mkfile(params [][]string) string {
	var path string
	var rFlag bool
	var size int64 = 0
	var cont string

	// Procesar parámetros
	for _, param := range params {
		if len(param) < 3 {
			continue
		}
		clave := param[1]
		valor := param[2]
		switch clave {
		case "path":
			path = valor
		case "r":
			rFlag = true
		case "size":
			fmt.Sscanf(valor, "%d", &size)
		case "cont":
			cont = valor
		}
	}

	if path == "" {
		return "❌ Error: El parámetro -path es obligatorio."
	}
	if size < 0 {
		return "❌ Error: El parámetro -size no puede ser negativo."
	}
	if cont != "" {
		if _, err := os.Stat(cont); os.IsNotExist(err) {
			return "❌ Error: El archivo especificado en -cont no existe."
		}
	}

	if !session.Sesion.LoggedIn {
		return "❌ Error: Debe iniciar sesión para crear archivos."
	}

	drive := string(session.Sesion.ID[0])
	dskPath := rutaBase + drive + ".dsk"
	file, err := os.OpenFile(dskPath, os.O_RDWR, 0666)
	if err != nil {
		return "❌ Error: disco no encontrado."
	}
	defer file.Close()

	// Leer MBR y superbloque
	var mbr structs.MBR
	binary.Read(file, binary.LittleEndian, &mbr)
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
	sb := structs.Superblock{}
	file.Seek(int64(part.PartStart), 0)
	binary.Read(file, binary.LittleEndian, &sb)

	// Procesar el path
	ruta := strings.Split(strings.Trim(path, "\"/"), "/")
	if len(ruta) < 1 {
		return "❌ Error: Ruta inválida."
	}
	nombreArchivo := ruta[len(ruta)-1]
	rutaCarpetas := ruta[:len(ruta)-1]

	// 1. Buscar/crear las carpetas padre
	inodoActual := int32(0) // raíz
	for _, nombreCarpeta := range rutaCarpetas {
		// Buscar si existe la carpeta en el FolderBlock del inodoActual
		encontrada := false
		var carpetaInodo structs.Inode
		file.Seek(int64(sb.SInodeStart)+int64(inodoActual)*int64(binary.Size(structs.Inode{})), 0)
		binary.Read(file, binary.LittleEndian, &carpetaInodo)
		for _, ptr := range carpetaInodo.IBlock {
			if ptr == -1 {
				continue
			}
			var folder structs.FolderBlock
			file.Seek(int64(sb.SBlockStart)+int64(ptr)*64, 0)
			binary.Read(file, binary.LittleEndian, &folder)
			for _, content := range folder.BContent {
				nombre := strings.Trim(string(content.BName[:]), "\x00")
				if nombre == nombreCarpeta && content.BInode != -1 {
					inodoActual = content.BInode
					encontrada = true
					break
				}
			}
			if encontrada {
				break
			}
		}
		if !encontrada {
			if !rFlag {
				return "❌ Error: Carpeta '" + nombreCarpeta + "' no existe. Usa -r para crearla."
			}
			// Crear la carpeta
			// Buscar inodo y bloque libres
			nuevoInodo := sb.SFirstInode
			nuevoBloque := sb.SFirstBlock
			// Crear inodo de carpeta
			inodoCarpeta := structs.Inode{
				IUid:  1,
				IGid:  1,
				ISize: 0,
				IType: 0, // carpeta
				IPerm: [3]byte{'7', '7', '7'},
			}
			for i := 0; i < 15; i++ {
				inodoCarpeta.IBlock[i] = -1
			}
			inodoCarpeta.IBlock[0] = nuevoBloque
			// Crear folder block
			var folder structs.FolderBlock
			copy(folder.BContent[0].BName[:], ".")
			folder.BContent[0].BInode = nuevoInodo
			copy(folder.BContent[1].BName[:], "..")
			folder.BContent[1].BInode = inodoActual
			for i := 2; i < 4; i++ {
				folder.BContent[i].BInode = -1
			}
			// Escribir inodo y bloque
			file.Seek(int64(sb.SInodeStart)+int64(nuevoInodo)*int64(binary.Size(structs.Inode{})), 0)
			binary.Write(file, binary.LittleEndian, &inodoCarpeta)
			file.Seek(int64(sb.SBlockStart)+int64(nuevoBloque)*64, 0)
			binary.Write(file, binary.LittleEndian, &folder)
			// Actualizar carpeta padre (agregar entrada)
			var folderPadre structs.FolderBlock
			file.Seek(int64(sb.SBlockStart)+int64(carpetaInodo.IBlock[0])*64, 0)
			binary.Read(file, binary.LittleEndian, &folderPadre)
			for i := 0; i < 4; i++ {
				if folderPadre.BContent[i].BInode == -1 {
					copy(folderPadre.BContent[i].BName[:], nombreCarpeta)
					folderPadre.BContent[i].BInode = nuevoInodo
					break
				}
			}
			file.Seek(int64(sb.SBlockStart)+int64(carpetaInodo.IBlock[0])*64, 0)
			binary.Write(file, binary.LittleEndian, &folderPadre)
			inodoActual = nuevoInodo
			// Actualizar superbloque, bitmaps, etc. (no implementado aquí)
			continue
		}
	}

	// 2. Verificar si el archivo ya existe en la carpeta destino
	var carpetaInodo structs.Inode
	file.Seek(int64(sb.SInodeStart)+int64(inodoActual)*int64(binary.Size(structs.Inode{})), 0)
	binary.Read(file, binary.LittleEndian, &carpetaInodo)
	for _, ptr := range carpetaInodo.IBlock {
		if ptr == -1 {
			continue
		}
		var folder structs.FolderBlock
		file.Seek(int64(sb.SBlockStart)+int64(ptr)*64, 0)
		binary.Read(file, binary.LittleEndian, &folder)
		for _, content := range folder.BContent {
			nombre := strings.Trim(string(content.BName[:]), "\x00")
			if nombre == nombreArchivo && content.BInode != -1 {
				return "❌ Error: El archivo ya existe."
			}
		}
	}

	// 3. Leer contenido a escribir
	var contenido []byte
	if cont != "" {
		contenido, _ = ioutil.ReadFile(cont)
	} else {
		contenido = []byte{}
	}
	if size > 0 {
		for int64(len(contenido)) < size {
			contenido = append(contenido, '0')
		}
		contenido = contenido[:size]
	}

	// 4. Buscar inodo y bloque libres (aquí solo el siguiente disponible)
	nuevoInodo := sb.SFirstInode
	nuevoBloque := sb.SFirstBlock

	// 5. Crear el inodo del archivo
	inodo := structs.Inode{
		IUid:  1, // UID del usuario en sesión (puedes buscarlo en users.txt)
		IGid:  1, // GID del grupo en sesión (puedes buscarlo en users.txt)
		ISize: int32(len(contenido)),
		IType: 1, // archivo
		IPerm: [3]byte{'6', '6', '4'},
	}
	for i := 0; i < 15; i++ {
		inodo.IBlock[i] = -1
	}
	inodo.IBlock[0] = nuevoBloque

	// 6. Crear el bloque de archivo
	var bloque structs.FileBlock
	copy(bloque.Content[:], contenido)

	// 7. Escribir el inodo y bloque en disco
	file.Seek(int64(sb.SInodeStart)+int64(nuevoInodo)*int64(binary.Size(structs.Inode{})), 0)
	binary.Write(file, binary.LittleEndian, &inodo)
	file.Seek(int64(sb.SBlockStart)+int64(nuevoBloque)*64, 0)
	binary.Write(file, binary.LittleEndian, &bloque)

	// 8. Actualizar el FolderBlock de la carpeta padre para agregar el archivo
	var folder structs.FolderBlock
	file.Seek(int64(sb.SBlockStart)+int64(carpetaInodo.IBlock[0])*64, 0)
	binary.Read(file, binary.LittleEndian, &folder)
	for i := 0; i < 4; i++ {
		if folder.BContent[i].BInode == -1 {
			copy(folder.BContent[i].BName[:], nombreArchivo)
			folder.BContent[i].BInode = nuevoInodo
			break
		}
	}
	file.Seek(int64(sb.SBlockStart)+int64(carpetaInodo.IBlock[0])*64, 0)
	binary.Write(file, binary.LittleEndian, &folder)

	// 9. Actualizar superbloque y bitmaps

	// --- Actualizar bitmap de inodos ---
	bitmapInodos := make([]byte, sb.SInodesCount)
	file.Seek(int64(sb.SBmInodeStart), 0)
	file.Read(bitmapInodos)
	bitmapInodos[nuevoInodo] = 1
	file.Seek(int64(sb.SBmInodeStart), 0)
	file.Write(bitmapInodos)

	// --- Actualizar bitmap de bloques ---
	bitmapBloques := make([]byte, sb.SBlocksCount)
	file.Seek(int64(sb.SBmBlockStart), 0)
	file.Read(bitmapBloques)
	bitmapBloques[nuevoBloque] = 1
	file.Seek(int64(sb.SBmBlockStart), 0)
	file.Write(bitmapBloques)

	// --- Actualizar superbloque ---
	sb.SFreeInodesCount--
	sb.SFreeBlocksCount--
	for i := nuevoInodo + 1; i < sb.SInodesCount; i++ {
		if bitmapInodos[i] == 0 {
			sb.SFirstInode = i
			break
		}
	}
	for i := nuevoBloque + 1; i < sb.SBlocksCount; i++ {
		if bitmapBloques[i] == 0 {
			sb.SFirstBlock = i
			break
		}
	}
	file.Seek(int64(part.PartStart), 0)
	binary.Write(file, binary.LittleEndian, &sb)

	return fmt.Sprintf("✅ Archivo %s creado correctamente en la partición %s", nombreArchivo, session.Sesion.ID)
}
