package diskmanager

import (
	"api-mia1/structs"
	"encoding/binary"
	"os"
	"strconv"
	"strings"
	"time"
)

func Mkfs(params [][]string) string {
	var id, fstype string
	// var tipo string = "full"

	for _, param := range params {
		key := strings.ToLower(param[1])
		val := strings.Trim(param[2], "\"")

		switch key {
		case "id":
			id = val
		case "type":
			if strings.ToLower(val) != "full" {
				return "❌ Error: solo se permite type=full"
			}
			// tipo = "full"
		case "fs":
			val = strings.ToLower(val)
			if val != "2fs" && val != "3fs" {
				return "❌ Error: fs debe ser 2fs o 3fs"
			}
			fstype = val
		default:
			return "⚠️ Parámetro no reconocido:" + key
		}
	}

	if id == "" {
		return "❌ Error: parámetro -id es obligatorio."
	}

	if fstype == "" {
		fstype = "2fs" // por defecto
	}

	output := formatearParticion(id, fstype)
	output += VerSuperblock(id)
	return output
}

func formatearParticion(id string, fstype string) string {
	// Obtener letra del disco
	drive := string(id[0])
	path := rutaBase + drive + ".dsk"

	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return "❌ Error: no se pudo abrir el disco."
	}
	defer file.Close()

	// Leer MBR
	var mbr structs.MBR
	if err := binary.Read(file, binary.LittleEndian, &mbr); err != nil {
		return "❌ Error al leer MBR."
	}

	// Buscar partición montada con ese ID
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
		return "❌ Error: no se encontró partición con el ID " + id
	}

	// Calcular n (cantidad de estructuras)
	superSize := int32(binary.Size(structs.Superblock{}))
	inodeSize := int32(binary.Size(structs.Inode{}))
	blockSize := int32(64) // según proyecto
	journalSize := int32(binary.Size(structs.Journaling{}))

	var n int32
	if fstype == "2fs" {
		// Fórmula: tamaño_particion = sizeOf(superblock) + n + 3*n + n*sizeOf(inodos) + 3*n*sizeOf(block)
		// tamaño_particion = sizeOf(superblock) + 4*n + n*sizeOf(inodos) + 3*n*sizeOf(block)
		// n = floor((part.PartSize - superSize) / (4 + inodeSize + 3*blockSize))
		n = (part.PartSize - superSize) / (4 + inodeSize + 3*blockSize)
	} else { // EXT3 incluye Journaling
		// Fórmula: tamaño_particion = sizeOf(superblock) + sizeOf(Journaling) + n + 3*n + n*sizeOf(inodos) + 3*n*sizeOf(block)
		// tamaño_particion = sizeOf(superblock) + sizeOf(Journaling) + 4*n + n*sizeOf(inodos) + 3*n*sizeOf(block)
		// n = floor((part.PartSize - superSize - journalSize) / (4 + inodeSize + 3*blockSize))
		n = (part.PartSize - superSize - journalSize) / (4 + inodeSize + 3*blockSize)
	}
	if n <= 0 {
		return "❌ Error: no hay espacio suficiente para formatear."
	}

	// Crear superbloque
	var sb structs.Superblock
	if fstype == "2fs" {
		sb.SFilesystemType = 2
	} else {
		sb.SFilesystemType = 3
	}
	sb.SInodesCount = n
	sb.SBlocksCount = 3 * n
	sb.SFreeInodesCount = n - 1
	sb.SFreeBlocksCount = 3*n - 2
	copy(sb.SMtime[:], time.Now().Format("2006-01-02 15:04:05 "))
	copy(sb.SUmountTime[:], time.Now().Format("2006-01-02 15:04:05 "))
	sb.SMagic = 0xEF53
	sb.SInodeSize = inodeSize
	sb.SBlockSize = blockSize
	sb.SFirstInode = 2
	sb.SFirstBlock = 2

	// Posiciones de inicio
	pos := part.PartStart + superSize
	if fstype == "3fs" {
		pos += journalSize // Saltar journaling
	}
	sb.SBmInodeStart = pos
	sb.SBmBlockStart = sb.SBmInodeStart + n
	sb.SInodeStart = sb.SBmBlockStart + 3*n
	sb.SBlockStart = sb.SInodeStart + inodeSize*n

	file.Seek(int64(part.PartStart), 0)
	binary.Write(file, binary.LittleEndian, &sb)

	// Inicializar Journaling para EXT3
	if fstype == "3fs" {
		file.Seek(int64(part.PartStart+superSize), 0)
		var journal structs.Journaling
		for i := int32(0); i < n; i++ {
			binary.Write(file, binary.LittleEndian, &journal)
		}
	}

	contenidoUsuarios := "1,G,root\n1,U,root,root,123\n"
	var bloqueUsers structs.Fileblock
	copy(bloqueUsers.BContent[:], contenidoUsuarios)

	inodeUsers := structs.Inode{
		IUid:  1,
		IGid:  1,
		ISize: int32(len(contenidoUsuarios)),
		IType: 1, // archivo
		IPerm: [3]byte{'7', '7', '7'},
	}
	copy(inodeUsers.ICtime[:], time.Now().Format("2006-01-02 15:04:05 "))
	copy(inodeUsers.IMtime[:], time.Now().Format("2006-01-02 15:04:05 "))
	copy(inodeUsers.IAtime[:], time.Now().Format("2006-01-02 15:04:05 "))
	inodeUsers.IBlock[0] = 0
	for i := 1; i < 15; i++ {
		inodeUsers.IBlock[i] = -1
	}

	inodeRoot := structs.Inode{
		IUid:  1,
		IGid:  1,
		ISize: 0,
		IType: 0, // carpeta
		IPerm: [3]byte{'7', '7', '7'},
	}
	copy(inodeRoot.ICtime[:], time.Now().Format("2006-01-02 15:04:05 "))
	copy(inodeRoot.IMtime[:], time.Now().Format("2006-01-02 15:04:05 "))
	copy(inodeRoot.IAtime[:], time.Now().Format("2006-01-02 15:04:05 "))
	inodeRoot.IBlock[0] = 1
	for i := 1; i < 15; i++ {
		inodeRoot.IBlock[i] = -1
	}

	carpeta := structs.FolderBlock{}
	copy(carpeta.BContent[0].BName[:], ".")
	carpeta.BContent[0].BInode = 0
	copy(carpeta.BContent[1].BName[:], "..")
	carpeta.BContent[1].BInode = 0
	copy(carpeta.BContent[2].BName[:], "users.txt")
	carpeta.BContent[2].BInode = 1
	carpeta.BContent[3].BInode = -1

	// Bitmap de inodos
	file.Seek(int64(sb.SBmInodeStart), 0)
	file.Write([]byte{1, 1})

	// Bitmap de bloques
	file.Seek(int64(sb.SBmBlockStart), 0)
	file.Write([]byte{1, 1})

	// Inodos
	file.Seek(int64(sb.SInodeStart), 0)
	binary.Write(file, binary.LittleEndian, &inodeRoot)
	binary.Write(file, binary.LittleEndian, &inodeUsers)

	// Bloques
	file.Seek(int64(sb.SBlockStart), 0)
	binary.Write(file, binary.LittleEndian, &bloqueUsers)
	binary.Write(file, binary.LittleEndian, &carpeta)

	sb.SFirstInode = 2
	sb.SFirstBlock = 2
	sb.SFreeInodesCount -= 2
	sb.SFreeBlocksCount -= 2

	file.Seek(int64(part.PartStart), 0)
	binary.Write(file, binary.LittleEndian, &sb)

	// Inicializar bitmaps e inodos básicos
	// Aquí escribirías el bitmap de inodos y bloques como 0s y marcos usados como 1s
	// Luego escribirías el inodo raíz y el archivo users.txt

	return "✅ Partición " + id + " formateada correctamente como " + fstype + "\n📦 Estado actual del MBR:\n" + mbr.PrintMBR(string(id[0]))
}

func VerSuperblock(id string) string {
	drive := string(id[0])
	path := rutaBase + drive + ".dsk"

	file, err := os.Open(path)
	if err != nil {
		return "❌ Error: no se pudo abrir el disco."
	}
	defer file.Close()

	var mbr structs.MBR
	if err := binary.Read(file, binary.LittleEndian, &mbr); err != nil {
		return "❌ Error al leer MBR."
	}

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
		return "❌ No se encontró partición con ID " + id
	}

	var sb structs.Superblock
	file.Seek(int64(part.PartStart), 0)
	if err := binary.Read(file, binary.LittleEndian, &sb); err != nil {
		return "❌ Error al leer el Superblock."
	}

	// Imprimir info básica
	output := ""
	output += "📦 SUPERBLOCK\n"
	output += "FilesystemType: " + strconv.Itoa(int(sb.SFilesystemType)) + "\n"
	output += "InodesCount: " + strconv.Itoa(int(sb.SInodesCount)) + "\n"
	output += "BlocksCount: " + strconv.Itoa(int(sb.SBlocksCount)) + "\n"
	output += "FreeInodesCount: " + strconv.Itoa(int(sb.SFreeInodesCount)) + "\n"
	output += "FreeBlocksCount: " + strconv.Itoa(int(sb.SFreeBlocksCount)) + "\n"
	output += "FirstInode: " + strconv.Itoa(int(sb.SFirstInode)) + "\n"
	output += "FirstBlock: " + strconv.Itoa(int(sb.SFirstBlock)) + "\n"
	output += "InodeStart: " + strconv.Itoa(int(sb.SInodeStart)) + "\n"
	output += "BlockStart: " + strconv.Itoa(int(sb.SBlockStart)) + "\n"
	return output
}
