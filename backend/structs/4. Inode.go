package structs

type Inode struct {
	IUid   int32
	IGid   int32
	ISize  int32
	IAtime [20]byte
	ICtime [20]byte
	IMtime [20]byte
	IBlock [15]int32
	IType  byte
	IPerm  [3]byte
}
