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
		return diskmanager.Logout()
	case "mkgrp":
		return diskmanager.Mkgrp(AnaliceRegExp(params))
	case "rmgrp":
		return diskmanager.Rmgrp(AnaliceRegExp(params))
	case "mkusr":
		return diskmanager.Mkusr(AnaliceRegExp(params))
	case "rmusr":
		return diskmanager.Rmusr(AnaliceRegExp(params))
	case "mkfile":
		return diskmanager.Mkfile(AnaliceRegExp(params))
	default:
		return ("Comando no reconocido: " + command)
	}
}
