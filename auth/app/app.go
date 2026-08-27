package app

import (
	"github.com/Crows-Storm/Axis/auth/app/command"
	"github.com/Crows-Storm/Axis/auth/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	RegisterUser command.RegisterUserCommandHandler
	Login        command.LoginCommandHandler
	//RegisterAndLogin command.RegisterAndLoginCommandHandler
	//RefreshToken     command.RefreshTokenCommandHandler
	Logout command.LogoutCommandHandler
}

type Queries struct {
	GetUser query.GetUserQueryHandler
	//GetPrincipal query.GetPrincipalHandler
}
