package structs

import "fmt"

type Partition struct {
	PartStatus      [1]byte
	PartType        [1]byte
	PartFit         [1]byte
	PartStart       int32
	PartSize        int32
	PartName        [16]byte
	PartCorrelative int32
	PartID          [4]byte
}

func (data Partition) PrintPartition() {
	fmt.Println(fmt.Sprintf("Name: %s, type: %s, start: %d, size: %d, status: %s, id: %s|", string(data.PartName[:]), string(data.PartType[:]), data.PartStart, data.PartSize, string(data.PartStatus[:]), string(data.PartID[:])))
}
