package diskmanager

import (
	"fmt"
	"proyecto1/session"
)

func Logout() {
	if session.Sesion.LoggedIn {
		fmt.Println("👋 Cerrando sesión del usuario:", session.Sesion.Username)
		session.Sesion = session.UsuarioActivo{} // reset
	} else {
		fmt.Println("⚠️ No hay ninguna sesión activa.")
	}
}
