package app

import "github.com/Crows-Storm/Axis/user/app/query"

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	//CreateUser command.CreateUserHandler
	//UpdateUser command.UpdateUserHandler
}

type Queries struct {
	GetUser query.GetUserQueryHandler
}
