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
	CreateUser       command.CreateUserCommandHandler
	CreateBatchUsers command.CreateBatchUserCommandHandler
	UpdateUser       command.UpdateUserCommandHandler
	SoftDeleteUser   command.SoftDeleteUserCommandHandler
	DisableUser      command.DisableUserCommandHandler
}

type Queries struct {
	GetUser            query.GetUserQueryHandler
	UserExists         query.UserExistsHandler
	UserStatusAnalysis query.UserStatusAnalysisHandler
	VerifyLogin        query.VerifyLoginHandler
}
