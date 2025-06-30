package analizador

import (
	"fmt"
	"os"
)

func AnalizerCommand(command string, params string) {
	switch command {
	case "mkdisk":
		fmt.Println("mkdisk")
	case "rmdisk":
		fmt.Println("rmdisk")
	case "fdisk":
		fmt.Println("fdisk")
	case "mount":
		fmt.Println("mount")
	case "unmount":
		fmt.Println("unmount")
	case "mkfs":
		fmt.Println("mkfs")
	case "login":
		fmt.Println("login")
	case "logout":
		fmt.Println("logout")
	case "mkgrp":
		fmt.Println("mkgrp")
	case "rmgrp":
		fmt.Println("rmgrp")
	case "mkusr":
		fmt.Println("mkusr")
	case "rmusr":
		fmt.Println("rmusr")
	case "cat":
		fmt.Println("cat")
	case "pause":
		fmt.Println("pause")
	case "execute":
		fmt.Println("execute")
	case "rep":
		fmt.Println("rep")
	case "exit":
		fmt.Println("Salir del sistema")
		os.Exit(0)
	default:
		fmt.Println("Comando no reconocido")
	}
}
