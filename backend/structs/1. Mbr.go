package structs

import (
	"fmt"
)

type MBR struct {
	MbrSize          int32
	MbrCreationDate  [20]byte
	MbrDiskSignature int32
	DiskFit          [1]byte
	Partitions       [4]Partition
}

func (m MBR) PrintMBR() {
	fmt.Println("---------- MBR ----------")
	fmt.Println(fmt.Sprintf("CreationDate: %s, fit: %s, size: %d", string(m.MbrCreationDate[:]), string(m.DiskFit[:]), m.MbrSize))
	for i := 0; i < 4; i++ {
		m.Partitions[i].PrintPartition()
	}
}
