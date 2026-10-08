package app

import (
	"github.com/Crows-Storm/Axis/audit/app/command"
	"github.com/Crows-Storm/Axis/audit/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	RecordEvent command.RecordAuditEventCommandHandler
}

type Queries struct {
	PageRoots     query.PageRootsQueryHandler
	GetEventGraph query.GetEventGraphQueryHandler
}
