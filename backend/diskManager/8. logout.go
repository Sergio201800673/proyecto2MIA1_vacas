package diskmanager

import (
	"api-mia1/session"
)

func Logout() string {
	if session.Sesion.LoggedIn {
		output := "👋 Cerrando sesión del usuario:" + session.Sesion.Username
		session.Sesion = session.UsuarioActivo{} // reset
		return output
	} else {
		return "⚠️ No hay ninguna sesión activa."
	}
}
