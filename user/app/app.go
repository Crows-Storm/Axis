package app

import (
	"github.com/Crows-Storm/Axis/user/app/command"
	"github.com/Crows-Storm/Axis/user/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	CreateUser command.CreateUserCommandHandler
	UpdateUser command.UpdateUserCommandHandler
}

type Queries struct {
	GetUser query.GetUserQueryHandler
}
