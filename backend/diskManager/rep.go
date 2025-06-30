package diskmanager

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"proyecto1/structs"
	"strconv"
	"strings"
	"unsafe"
)

func Rep(params [][]string) {
	var name, path, id, ruta string

	// Extraer parámetros
	for _, param := range params {
		key := strings.ToLower(param[1])
		val := strings.Trim(param[2], "\"")

		switch key {
		case "name":
			name = val
		case "path":
			path = val
		case "id":
			id = val
		case "ruta":
			ruta = val
		default:
			fmt.Println("⚠️ Parámetro no reconocido:", key)
		}
	}

	// Validar parámetros obligatorios
	if name == "" || path == "" || id == "" {
		fmt.Println("❌ Error: parámetros -name, -path e -id son obligatorios")
		return
	}

	// Obtener información del disco y partición
	drive := string(id[0])
	diskPath := rutaBase + drive + ".dsk"

	file, err := os.OpenFile(diskPath, os.O_RDWR, 0666)
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

	// Leer superbloque
	var sb structs.Superblock
	file.Seek(int64(part.PartStart), 0)
	if err := binary.Read(file, binary.LittleEndian, &sb); err != nil {
		fmt.Println("❌ Error al leer el Superblock.")
		return
	}

	// Generar reporte según el tipo
	switch name {
	case "mbr":
		generateMBRReport(mbr, path)
	case "disk":
		generateDiskReport(mbr, path, part)
	case "inode":
		generateInodeReport(sb, file, path)
	case "block":
		generateBlockReport(sb, file, path)
	case "bm_inode":
		generateBmInodeReport(sb, file, path)
	case "bm_block":
		generateBmBlockReport(sb, file, path)
	case "tree":
		generateTreeReport(sb, file, path)
	case "sb":
		generateSuperblockReport(sb, path)
	case "file":
		if ruta == "" {
			fmt.Println("❌ Error: parámetro -ruta es obligatorio para reporte file")
			return
		}
		generateFileReport(sb, file, path, ruta)
	case "ls":
		if ruta == "" {
			fmt.Println("❌ Error: parámetro -ruta es obligatorio para reporte ls")
			return
		}
		generateLsReport(sb, file, path, ruta)
	/* case "journaling":
	if sb.SFilesystemType != 3 {
		fmt.Println("❌ Error: journaling solo disponible en sistemas EXT3")
		return
	}
	generateJournalingReport(sb, file, path) */
	default:
		fmt.Println("❌ Error: tipo de reporte no válido:", name)
	}
}

/* Funciona pero hay algo escrito que no genera el archivo final */
func generateMBRReport(mbr structs.MBR, outputPath string) {
	// Crear archivo DOT
	dotContent := `digraph MBR {
    node [shape=plaintext]
    mbr [label=<
    <table border="1" cellborder="1" cellspacing="0">
        <tr><td colspan="2"><b>REPORTE DE MBR</b></td></tr>
        <tr><td>mbr_tamano</td><td>` + fmt.Sprint(mbr.MbrSize) + `</td></tr>
        <tr><td>mbr_fecha_creacion</td><td>` + string(mbr.MbrCreationDate[:]) + `</td></tr>
        <tr><td>mbr_disk_signature</td><td>` + fmt.Sprint(mbr.MbrDiskSignature) + `</td></tr>
        <tr><td>dsk_fit</td><td>` + string(mbr.DiskFit[:]) + `</td></tr>`

	// Agregar información de particiones
	for i, p := range mbr.Partitions {
		if p.PartStatus[0] != '0' {
			dotContent += `
        <tr><td colspan="2"><b>Particion ` + strconv.Itoa(i+1) + `</b></td></tr>
        <tr><td>part_status</td><td>` + string(p.PartStatus[:]) + `</td></tr>
        <tr><td>part_type</td><td>` + string(p.PartType[:]) + `</td></tr>
        <tr><td>part_fit</td><td>` + string(p.PartFit[:]) + `</td></tr>
        <tr><td>part_start</td><td>` + fmt.Sprint(p.PartStart) + `</td></tr>
        <tr><td>part_size</td><td>` + fmt.Sprint(p.PartSize) + `</td></tr>
        <tr><td>part_name</td><td>` + string(p.PartName[:]) + `</td></tr>`
		}
	}

	dotContent += `
    </table>>];
}`

	// Generar imagen con Graphviz
	fmt.Println(dotContent)
	generateGraphvizImage(dotContent, outputPath)
	fmt.Println("✅ Reporte MBR generado en:", outputPath)
}

/* Funciona correcto*/
func generateDiskReport(mbr structs.MBR, outputPath string, part structs.Partition) {
	totalSize := int(mbr.MbrSize)
	usedSpace := 0
	var partitions []struct {
		name  string
		size  int
		start int
	}

	// Calcular espacio usado y obtener particiones
	for _, p := range mbr.Partitions {
		if p.PartStatus[0] != '0' {
			usedSpace += int(p.PartSize)
			partitions = append(partitions, struct {
				name  string
				size  int
				start int
			}{
				name:  strings.Trim(string(p.PartName[:]), "\x00"),
				size:  int(p.PartSize),
				start: int(p.PartStart),
			})
		}
	}
	freeSpace := totalSize - usedSpace

	// Crear archivo DOT
	dotContent := `digraph DISK {
    node [shape=plaintext]
    disk [label=<
    <table border="1" cellborder="1" cellspacing="0">
        <tr><td colspan="3"><b>REPORTE DE DISK</b></td></tr>
        <tr><td><b>Nombre</b></td><td><b>Inicio</b></td><td><b>Tamaño</b></td></tr>`

	// Agregar particiones
	for _, p := range partitions {
		percentage := float64(p.size) / float64(totalSize) * 100
		dotContent += `
        <tr><td>` + p.name + `</td><td>` + fmt.Sprint(p.start) + `</td><td>` + fmt.Sprintf("%.2f%%", percentage) + `</td></tr>`
	}

	// Agregar espacio libre
	freePercentage := float64(freeSpace) / float64(totalSize) * 100
	dotContent += `
        <tr><td>Libre</td><td>-</td><td>` + fmt.Sprintf("%.2f%%", freePercentage) + `</td></tr>
    </table>>];
}`

	generateGraphvizImage(dotContent, outputPath)
	fmt.Println("✅ Reporte DISK generado en:", outputPath)
}

/* Funciona correto */
func generateInodeReport(sb structs.Superblock, file *os.File, outputPath string) {
	// 1. Leer el primer inodo (o el inodo que corresponda según tu lógica)
	//    (Este es un ejemplo, ajusta según cómo se organice tu sistema de archivos)
	var inode structs.Inode
	inodeStart := sb.SFirstInode // Posición del primer inodo
	file.Seek(int64(inodeStart), 0)
	if err := binary.Read(file, binary.LittleEndian, &inode); err != nil {
		fmt.Println("❌ Error al leer el Inodo:", err)
		return
	}

	// 2. Función auxiliar para limpiar campos de tiempo
	cleanText := func(data []byte) string {
		str := string(data[:])
		str = strings.TrimSpace(str)
		return strings.ReplaceAll(str, " ", "**")
	}

	// 3. Generar el contenido DOT
	dotContent := `digraph INODE {
    node [shape=plaintext]
    inode [label=<
    <table border="1" cellborder="1" cellspacing="0">
        <tr><td colspan="2"><b>REPORTE DE INODO</b></td></tr>
        <tr><td>i_uid</td><td>` + fmt.Sprint(inode.IUid) + `</td></tr>
        <tr><td>i_gid</td><td>` + fmt.Sprint(inode.IGid) + `</td></tr>
        <tr><td>i_size</td><td>` + fmt.Sprint(inode.ISize) + ` bytes</td></tr>
        <tr><td>i_atime</td><td>` + cleanText(inode.IAtime[:]) + `</td></tr>
        <tr><td>i_ctime</td><td>` + cleanText(inode.ICtime[:]) + `</td></tr>
        <tr><td>i_mtime</td><td>` + cleanText(inode.IMtime[:]) + `</td></tr>
        <tr><td>i_type</td><td>` + string(inode.IType) + `</td></tr>
        <tr><td>i_perm</td><td>` + fmt.Sprintf("%o", inode.IPerm) + `</td></tr>
        <tr><td>i_block</td><td>` + formatBlocks(inode.IBlock) + `</td></tr>
    </table>>];
}`

	// 4. Generar la imagen
	generateGraphvizImage(dotContent, outputPath)
	fmt.Println("✅ Reporte de INODO generado en:", outputPath)
}

/* Función auxiliar para formatear bloques *** aun de generateInodeReport */
func formatBlocks(blocks [15]int32) string {
	var sb strings.Builder
	sb.WriteString("[")
	for i, block := range blocks {
		if block != -1 {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprint(block))
		}
	}
	sb.WriteString("]")
	return sb.String()
}

/* No hace nada xd */
func generateBlockReport(sb structs.Superblock, file *os.File, outputPath string) {
	// Implementar lógica para leer bloques y generar reporte
	fmt.Println("✅ Reporte BLOCK generado en:", outputPath)
}

/* No se si funciona bien ;( */
func generateBmInodeReport(sb structs.Superblock, file *os.File, outputPath string) {
	// Leer bitmap de inodos
	file.Seek(int64(sb.SBmInodeStart), 0)
	bitmap := make([]byte, sb.SInodesCount)
	file.Read(bitmap)

	// Crear archivo de texto
	content := "BITMAP DE INODOS\n"
	for i, b := range bitmap {
		if i > 0 && i%20 == 0 {
			content += "\n"
		}
		content += fmt.Sprintf("%d ", b)
	}

	err := os.WriteFile(outputPath, []byte(content), 0644)
	if err != nil {
		fmt.Println("❌ Error al generar reporte bm_inode:", err)
		return
	}
	fmt.Println("✅ Reporte BM_INODE generado en:", outputPath)
}

/* No llega */
func generateBmBlockReport(sb structs.Superblock, file *os.File, outputPath string) {
	// Leer bitmap de bloques
	file.Seek(int64(sb.SBmBlockStart), 0)
	bitmap := make([]byte, sb.SBlocksCount)
	file.Read(bitmap)

	// Crear archivo de texto
	content := "BITMAP DE BLOQUES\n"
	for i, b := range bitmap {
		if i > 0 && i%20 == 0 {
			content += "\n"
		}
		content += fmt.Sprintf("%d ", b)
	}

	err := os.WriteFile(outputPath, []byte(content), 0644)
	if err != nil {
		fmt.Println("❌ Error al generar reporte bm_block:", err)
		return
	}
	fmt.Println("✅ Reporte BM_BLOCK generado en:", outputPath)
}

/* No hace nada xd */
func generateTreeReport(sb structs.Superblock, file *os.File, outputPath string) {
	// Funciones auxiliares definidas primero
	colorForInode := func(inode structs.Inode) string {
		if inode.IType == '0' {
			return "#F0F8FF" // Archivo: AliceBlue
		}
		return "#FFFACD" // Directorio: LemonChiffon
	}

	colorForBlock := func(block structs.Block) string {
		switch block.BType {
		case '0':
			return "#E6E6FA" // Folder: Lavender
		case '1':
			return "#98FB98" // File: PaleGreen
		case '2':
			return "#FFB6C1" // Pointer: LightPink
		default:
			return "#FFFFFF"
		}
	}

	shortenText := func(text string, maxLen int) string {
		if len(text) > maxLen {
			return text[:maxLen-3] + "..."
		}
		return text
	}

	formatName := func(name [12]byte) string {
		return strings.TrimRight(string(name[:]), "\x00")
	}

	// Inicio del contenido DOT
	var dotBuilder strings.Builder
	dotBuilder.WriteString(`digraph TREE {
    node [shape=plaintext]
    rankdir=LR
    nodesep=0.5

    edge [arrowhead=none]
`)

	processedInodes := make(map[int32]bool)
	processedBlocks := make(map[int32]bool)

	// Función recursiva para procesar inodos
	var processInode func(inodeNum int32, isRoot bool)
	processInode = func(inodeNum int32, isRoot bool) {
		if inodeNum == -1 || processedInodes[inodeNum] {
			return
		}
		processedInodes[inodeNum] = true

		// Leer inodo
		inodePos := sb.SFirstInode + inodeNum*int32(unsafe.Sizeof(structs.Inode{}))
		var inode structs.Inode
		file.Seek(int64(inodePos), 0)
		if err := binary.Read(file, binary.LittleEndian, &inode); err != nil {
			fmt.Printf("❌ Error leyendo inodo %d: %v\n", inodeNum, err)
			return
		}

		// Nodo del inodo
		dotBuilder.WriteString(fmt.Sprintf(`
    inode_%d [label=<
        <table border="1" cellborder="1" cellspacing="0" bgcolor="%s">
            <tr><td colspan="2"><b>Inodo %d</b></td></tr>
            <tr><td>Tipo</td><td>%c</td></tr>
            <tr><td>Permisos</td><td>%o</td></tr>
            <tr><td>Tamaño</td><td>%d bytes</td></tr>
        </table>
    >];`, inodeNum, colorForInode(inode), inodeNum, inode.IType, inode.IPerm, inode.ISize))

		// Procesar bloques
		for i, blockPtr := range inode.IBlock {
			if blockPtr == -1 {
				continue
			}

			if processedBlocks[blockPtr] {
				dotBuilder.WriteString(fmt.Sprintf("\n    inode_%d -> block_%d [label=\"%d\"];",
					inodeNum, blockPtr, i))
				continue
			}
			processedBlocks[blockPtr] = true

			// Leer bloque
			blockPos := sb.SFirstBlock + blockPtr*int32(sb.SBlockSize)
			var block structs.Block
			file.Seek(int64(blockPos), 0)
			if err := binary.Read(file, binary.LittleEndian, &block); err != nil {
				fmt.Printf("❌ Error leyendo bloque %d: %v\n", blockPtr, err)
				continue
			}

			// Nodo del bloque
			dotBuilder.WriteString(fmt.Sprintf(`
    block_%d [label=<
        <table border="1" cellborder="1" cellspacing="0" bgcolor="%s">
            <tr><td colspan="2"><b>Bloque %d</b></td></tr>
            <tr><td>Tipo</td><td>%c</td></tr>`,
				blockPtr, colorForBlock(block), blockPtr, block.BType))

			switch block.BType {
			case '0': // FolderBlock
				folder := block.BContent.(structs.FolderBlock)
				dotBuilder.WriteString(`<tr><td colspan="2"><table border="0" cellborder="1">`)
				for _, content := range folder.BContent {
					if content.BInode != -1 {
						dotBuilder.WriteString(fmt.Sprintf(
							`<tr><td align="left">%s</td><td align="right">Ino:%d</td></tr>`,
							formatName(content.BName), content.BInode))
					}
				}
				dotBuilder.WriteString(`</table></td></tr>`)

			case '1': // FileBlock
				content := block.BContent.(structs.FileBlock)
				text := strings.TrimRight(string(content.Content[:]), "\x00")
				dotBuilder.WriteString(fmt.Sprintf(
					`<tr><td colspan="2"><font face="Courier">%s</font></td></tr>`,
					template.HTMLEscapeString(shortenText(text, 30))))

			case '2': // PointerBlock
				pointers := block.BContent.(structs.PointerBlock)
				dotBuilder.WriteString(`<tr><td colspan="2"><table border="0" cellborder="1">`)
				for i, ptr := range pointers.BPointer {
					if ptr != -1 {
						dotBuilder.WriteString(fmt.Sprintf(
							`<tr><td align="left">Ptr%d</td><td align="right">→%d</td></tr>`,
							i, ptr))
					}
				}
				dotBuilder.WriteString(`</table></td></tr>`)
			}

			dotBuilder.WriteString("</table>>];\n")
			dotBuilder.WriteString(fmt.Sprintf("    inode_%d -> block_%d [label=\"%d\"];",
				inodeNum, blockPtr, i))

			// Procesar bloques apuntadores recursivamente
			if block.BType == '2' {
				pointers := block.BContent.(structs.PointerBlock)
				for _, ptr := range pointers.BPointer {
					if ptr != -1 {
						processInode(ptr, false)
					}
				}
			}
		}
	}

	// Procesar desde el inodo raíz
	processInode(0, true)
	dotBuilder.WriteString("\n}")

	generateGraphvizImage(dotBuilder.String(), outputPath)
	fmt.Println("✅ Reporte TREE generado en:", outputPath)
}

/* Funciona pero tiene un espacio que hace que no funcione jajaja */
func generateSuperblockReport(sb structs.Superblock, outputPath string) {

	cleanTime := func(timeData []byte) string {
		strTime := string(timeData[:])
		// Elimina espacios al inicio y final, luego reemplaza espacios internos con **
		strTime = strings.TrimSpace(strTime)
		return strings.ReplaceAll(strTime, " ", "**")
	}

	dotContent := `digraph SUPERBLOCK {
    node [shape=plaintext]
    sb [label=<
    <table border="1" cellborder="1" cellspacing="0">
        <tr><td colspan="2"><b>REPORTE DE SUPERBLOQUE</b></td></tr>
        <tr><td>s_filesystem_type</td><td>` + fmt.Sprint(sb.SFilesystemType) + `</td></tr>
        <tr><td>s_inodes_count</td><td>` + fmt.Sprint(sb.SInodesCount) + `</td></tr>
        <tr><td>s_blocks_count</td><td>` + fmt.Sprint(sb.SBlocksCount) + `</td></tr>
        <tr><td>s_free_inodes_count</td><td>` + fmt.Sprint(sb.SFreeInodesCount) + `</td></tr>
        <tr><td>s_free_blocks_count</td><td>` + fmt.Sprint(sb.SFreeBlocksCount) + `</td></tr>
        <tr><td>s_mtime</td><td>` + cleanTime(sb.SMtime[:]) + `</td></tr>
        <tr><td>s_umtime</td><td>` + cleanTime(sb.SUmountTime[:]) + `</td></tr>
        <tr><td>s_magic</td><td>` + fmt.Sprintf("0x%X", sb.SMagic) + `</td></tr>
        <tr><td>s_inode_s</td><td>` + fmt.Sprint(sb.SInodeSize) + `</td></tr>
        <tr><td>s_block_s</td><td>` + fmt.Sprint(sb.SBlockSize) + `</td></tr>
        <tr><td>s_first_ino</td><td>` + fmt.Sprint(sb.SFirstInode) + `</td></tr>
        <tr><td>s_first_blo</td><td>` + fmt.Sprint(sb.SFirstBlock) + `</td></tr>
    </table>>];
}`

	generateGraphvizImage(dotContent, outputPath)
	fmt.Println("✅ Reporte SUPERBLOCK generado en:", outputPath)
}

/* No hace nada xd */
func generateFileReport(sb structs.Superblock, file *os.File, outputPath, filePath string) {
	// Implementar lógica para leer archivo y generar reporte
	fmt.Println("✅ Reporte FILE generado en:", outputPath)
}

/* No hace nada xd */
func generateLsReport(sb structs.Superblock, file *os.File, outputPath, dirPath string) {
	// Implementar lógica para listar directorio y generar reporte
	fmt.Println("✅ Reporte LS generado en:", outputPath)
}

/* func generateJournalingReport(sb structs.Superblock, file *os.File, outputPath string, part structs.Partition) {
	// Leer journaling (solo para EXT3)
	journalStart := part.PartStart + int32(binary.Size(structs.Superblock{}))
	journalSize := sb.SInodesCount // Asumiendo que journaling tiene tamaño n

	file.Seek(int64(journalStart), 0)
	journals := make([]structs.Journal, journalSize)
	binary.Read(file, binary.LittleEndian, &journals)

	// Crear archivo DOT
	dotContent := `digraph JOURNALING {
    node [shape=plaintext]
    journal [label=<
    <table border="1" cellborder="1" cellspacing="0">
        <tr><td colspan="4"><b>REPORTE DE JOURNALING</b></td></tr>
        <tr><td><b>Operación</b></td><td><b>Path</b></td><td><b>Contenido</b></td><td><b>Fecha</b></td></tr>`

	for _, j := range journals {
		if j.Operation[0] != 0 {
			dotContent += `
        <tr>
            <td>` + string(j.Operation[:]) + `</td>
            <td>` + string(j.Path[:]) + `</td>
            <td>` + string(j.Content[:]) + `</td>
            <td>` + string(j.Date[:]) + `</td>
        </tr>`
		}
	}

	dotContent += `
    </table>>];
}`

	generateGraphvizImage(dotContent, outputPath)
	fmt.Println("✅ Reporte JOURNALING generado en:", outputPath)
} */

func generateGraphvizImage(dotContent string, outputPath string) {
	// Crear directorio si no existe
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Println("❌ Error al crear directorio:", err)
		return
	}

	// Cambiar extensión a .dot para el archivo intermedio
	dotPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".dot"

	// 1. Crear archivo DOT en la carpeta de destino
	if err := os.WriteFile(dotPath, []byte(dotContent), 0644); err != nil {
		fmt.Println("❌ Error al crear archivo DOT:", err)
		return
	}
	fmt.Println("✅ Archivo DOT generado en:", dotPath)

	// 2. Convertir DOT a PNG (u otro formato)
	format := "png" // Usaremos PNG por defecto

	// Si la salida original no es PNG, mantenemos el formato solicitado
	ext := strings.ToLower(filepath.Ext(outputPath))
	if ext == ".jpg" || ext == ".jpeg" {
		format = "jpeg"
	} else if ext == ".pdf" {
		format = "pdf"
	} else if ext == ".svg" {
		format = "svg"
	}

	// Verificar que Graphviz está instalado
	if _, err := exec.LookPath("dot"); err != nil {
		fmt.Println("❌ Graphviz no está instalado o no está en el PATH")
		fmt.Println("   Instala con: sudo pacman -S graphviz")
		return
	}

	// Ejecutar conversión
	cmd := exec.Command("dot", "-T"+format, dotPath, "-o", outputPath)

	// Capturar salida de error
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("❌ Error al convertir DOT a imagen:")
		fmt.Println("   - Error:", err)
		fmt.Println("   - Detalles:", stderr.String())
		return
	}
	fmt.Println(dotContent)
	fmt.Println("✅ Reporte generado correctamente en:", outputPath)
}
