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

func (m MBR) PrintMBR(diskName string) string {

	var output string = ""
	output += "---------- MBR ----------\n"
	output += fmt.Sprintf("Disco: "+diskName+"\nCreationDate: %s, fit: %s, size: %d\n", string(m.MbrCreationDate[:]), string(m.DiskFit[:]), m.MbrSize)
	for i := 0; i < 4; i++ {
		output += m.Partitions[i].PrintPartition()
	}
	return output
}
