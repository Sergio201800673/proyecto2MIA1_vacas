package diskmanager

import (
	"encoding/binary"
	"os"
)

func binaryWrite(file *os.File, data interface{}) error {
	err := binary.Write(file, binary.LittleEndian, data)
	return err
}
