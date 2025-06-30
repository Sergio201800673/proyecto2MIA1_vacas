package structs

type FolderBlock struct {
	BContent [4]Content
}

type Content struct {
	BName  [12]byte
	BInode int32
}

type Block struct {
	BType    byte
	BContent interface{}
}

type FileBlock struct {
	Content [64]byte
}
