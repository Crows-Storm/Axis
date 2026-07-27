package app

import "github.com/Crows-Storm/Axis/auth/app/command"

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	RegisterUser command.RegisterUserCommandHandler
}

type Queries struct {
	//GetPrincipal query.GetPrincipalHandler
}
