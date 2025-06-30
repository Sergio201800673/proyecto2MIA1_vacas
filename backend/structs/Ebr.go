package structs

type EBR struct {
	PartStatus [1]byte
	PartFit    [1]byte
	PartStart  int32
	PartSize   int32
	PartNext   int32
	PartName   [16]byte
}
