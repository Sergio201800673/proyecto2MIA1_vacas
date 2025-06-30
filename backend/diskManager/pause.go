package diskmanager

import (
	"bufio"
	"fmt"
	"os"
)

func Pause() {
	fmt.Println("⏸️ Presiona Enter para continuar...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
