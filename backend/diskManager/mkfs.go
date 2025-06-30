package diskmanager

import (
	"encoding/binary"
	"fmt"
	"os"
	"proyecto1/structs"
	"strings"
	"time"
)

func Mkfs(params [][]string) {
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
				fmt.Println("❌ Error: solo se permite type=full")
				return
			}
			// tipo = "full"
		case "fs":
			val = strings.ToLower(val)
			if val != "2fs" && val != "3fs" {
				fmt.Println("❌ Error: fs debe ser 2fs o 3fs")
				return
			}
			fstype = val
		default:
			fmt.Println("⚠️ Parámetro no reconocido:", key)
		}
	}

	if id == "" {
		fmt.Println("❌ Error: parámetro -id es obligatorio.")
		return
	}

	if fstype == "" {
		fstype = "2fs" // por defecto
	}

	formatearParticion(id, fstype)
	VerSuperblock(id)
}

func formatearParticion(id string, fstype string) {
	// Obtener letra del disco
	drive := string(id[0])
	path := rutaBase + drive + ".dsk"

	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("❌ Error: no se pudo abrir el disco.")
		return
	}
	defer file.Close()

	// Leer MBR
	var mbr structs.MBR
	if err := binary.Read(file, binary.LittleEndian, &mbr); err != nil {
		fmt.Println("❌ Error al leer MBR.")
		return
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
		fmt.Println("❌ Error: no se encontró partición con el ID", id)
		return
	}

	// Calcular n (cantidad de estructuras)
	var n int32
	superSize := int32(binary.Size(structs.Superblock{}))
	inodeSize := int32(binary.Size(structs.Inode{}))
	blockSize := int32(64) // según proyecto

	// Cálculo basado en tipo EXT2 o EXT3
	if fstype == "2fs" {
		n = (part.PartSize - superSize) / (4 + inodeSize + 3*blockSize + 3)
	} else { // EXT3 incluye Journaling
		journalSize := int32(binary.Size(structs.Journaling{}))
		n = (part.PartSize - superSize) / (4 + journalSize + inodeSize + 3*blockSize + 4)
	}
	if n <= 0 {
		fmt.Println("❌ Error: no hay espacio suficiente para formatear.")
		return
	}

	// Crear superbloque
	var sb structs.Superblock
	sb.SFilesystemType = 2
	if fstype == "3fs" {
		sb.SFilesystemType = 3
	}
	sb.SInodesCount = n
	sb.SBlocksCount = 3 * n
	sb.SFreeInodesCount = n - 1
	sb.SFreeBlocksCount = 3*n - 2 // porque usaremos 2 bloques para / y users.txt
	copy(sb.SMtime[:], time.Now().Format("2006-01-02 15:04:05"))
	copy(sb.SUmountTime[:], time.Now().Format("2006-01-02 15:04:05"))
	sb.SMagic = 0xEF53
	sb.SInodeSize = inodeSize
	sb.SBlockSize = blockSize
	sb.SFirstInode = 2
	sb.SFirstBlock = 2
	sb.SBmInodeStart = part.PartStart + superSize
	sb.SBmBlockStart = sb.SBmInodeStart + n
	sb.SInodeStart = sb.SBmBlockStart + 3*n
	sb.SBlockStart = sb.SInodeStart + inodeSize*n

	file.Seek(int64(part.PartStart), 0)
	binary.Write(file, binary.LittleEndian, &sb)

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
	copy(inodeUsers.ICtime[:], time.Now().Format("2006-01-02 15:04:05"))
	copy(inodeUsers.IMtime[:], time.Now().Format("2006-01-02 15:04:05"))
	copy(inodeUsers.IAtime[:], time.Now().Format("2006-01-02 15:04:05"))
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
	copy(inodeRoot.ICtime[:], time.Now().Format("2006-01-02 15:04:05"))
	copy(inodeRoot.IMtime[:], time.Now().Format("2006-01-02 15:04:05"))
	copy(inodeRoot.IAtime[:], time.Now().Format("2006-01-02 15:04:05"))
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

	fmt.Println("✅ Partición", id, "formateada correctamente como", fstype)
	mbr.PrintMBR()
}

func VerSuperblock(id string) {
	drive := string(id[0])
	path := rutaBase + drive + ".dsk"

	file, err := os.Open(path)
	if err != nil {
		fmt.Println("❌ Error: no se pudo abrir el disco.")
		return
	}
	defer file.Close()

	var mbr structs.MBR
	if err := binary.Read(file, binary.LittleEndian, &mbr); err != nil {
		fmt.Println("❌ Error al leer MBR.")
		return
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
		fmt.Println("❌ No se encontró partición con ID", id)
		return
	}

	var sb structs.Superblock
	file.Seek(int64(part.PartStart), 0)
	if err := binary.Read(file, binary.LittleEndian, &sb); err != nil {
		fmt.Println("❌ Error al leer el Superblock.")
		return
	}

	// Imprimir info básica
	fmt.Println("📦 SUPERBLOCK")
	fmt.Println("FilesystemType:", sb.SFilesystemType)
	fmt.Println("InodesCount:", sb.SInodesCount)
	fmt.Println("BlocksCount:", sb.SBlocksCount)
	fmt.Println("FreeInodesCount:", sb.SFreeInodesCount)
	fmt.Println("FreeBlocksCount:", sb.SFreeBlocksCount)
	fmt.Println("FirstInode:", sb.SFirstInode)
	fmt.Println("FirstBlock:", sb.SFirstBlock)
	fmt.Println("InodeStart:", sb.SInodeStart)
	fmt.Println("BlockStart:", sb.SBlockStart)
}
