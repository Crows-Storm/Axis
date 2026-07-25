package app

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	//CreateUser command.CreateUserHandler
}

type Queries struct {
	//GetPrincipal query.GetPrincipalHandler
}
