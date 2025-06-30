package structs

type Superblock struct {
	SFilesystemType  int32
	SInodesCount     int32
	SBlocksCount     int32
	SFreeBlocksCount int32
	SFreeInodesCount int32
	SMtime           [20]byte
	SUmountTime      [20]byte
	SMntCount        int32
	SMagic           int32
	SInodeSize       int32
	SBlockSize       int32
	SFirstInode      int32
	SFirstBlock      int32
	SBmInodeStart    int32
	SBmBlockStart    int32
	SInodeStart      int32
	SBlockStart      int32
}
