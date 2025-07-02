package session

type UsuarioActivo struct {
	Username string
	Group    string
	ID       string
	LoggedIn bool
}

var Sesion UsuarioActivo
