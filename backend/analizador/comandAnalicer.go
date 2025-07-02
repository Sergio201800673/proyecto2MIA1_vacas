package analizador

import (
	diskmanager "api-mia1/diskManager"
)

func AnalizerCommand(command string, params string) string {
	switch command {
	case "mkdisk":
		return diskmanager.Mkdisk(AnaliceRegExp(params))
	case "rmdisk":
		return diskmanager.Rmdisk(AnaliceRegExp(params))
	case "fdisk":
		return diskmanager.Fdisk(AnaliceRegExp(params))
	case "mount":
		return diskmanager.Mount(AnaliceRegExp(params))
	case "unmount":
		return diskmanager.Unmount(AnaliceRegExp(params))
	case "mkfs":
		return diskmanager.Mkfs(AnaliceRegExp(params))
	case "login":
		return diskmanager.Login(AnaliceRegExp(params))
	case "logout":
		return diskmanager.Logout(AnaliceRegExp(params))
	case "mkgrp":
		return "mkgrp"
	case "rmgrp":
		return "rmgrp"
	case "mkusr":
		return "mkusr"
	case "rmusr":
		return "rmusr"
	case "cat":
		return "cat"
		/* 	case "pause":
		return "pause" */
	case "rep":
		return "rep"
	default:
		return ("Comando no reconocido: " + command)
	}
}
