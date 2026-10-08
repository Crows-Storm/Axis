package main

import (
	"errors"
	"strings"

	"github.com/Crows-Storm/Axis/audit/app"
	"github.com/Crows-Storm/Axis/audit/app/query"
	"github.com/Crows-Storm/Axis/audit/domain/audit"
	"github.com/Crows-Storm/Axis/audit/protos"
	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/pagination"
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	app app.Application
}

var _ protos.ServerInterface = (*HTTPServer)(nil)

func (h HTTPServer) PageRootEvents(c *gin.Context, params protos.PageRootEventsParams) {
	q := query.PageRootsQuery{
		Args: pagination.Args{
			First:  params.First,
			After:  params.After,
			Last:   params.Last,
			Before: params.Before,
		},
	}
	if params.Service != nil {
		q.Service = strings.TrimSpace(*params.Service)
	}
	if params.EventName != nil {
		q.EventName = strings.TrimSpace(*params.EventName)
	}
	if params.From != nil {
		from := params.From.UTC()
		q.From = &from
	}
	if params.To != nil {
		to := params.To.UTC()
		q.To = &to
	}

	rows, err := h.app.Queries.PageRoots.Handle(c.Request.Context(), q)
	if err != nil {
		writeAuditError(c, err)
		return
	}
	server.Success(c, rows)
}

func (h HTTPServer) GetEventGraph(c *gin.Context, eventId string) {
	graph, err := h.app.Queries.GetEventGraph.Handle(c.Request.Context(), query.GetEventGraphQuery{EventID: eventId})
	if err != nil {
		writeAuditError(c, err)
		return
	}
	server.Success(c, graph)
}

func writeAuditError(c *gin.Context, err error) {
	var executed decorator.CommandExecutedError
	if errors.As(err, &executed) || errors.Is(err, audit.ErrInvalidRecord) {
		server.ErrorWithCodeAndMessage(c, server.CodeBadRequest, err.Error())
		return
	}
	if errors.Is(err, pagination.ErrInvalidArgs) ||
		errors.Is(err, pagination.ErrFirstLastConflict) ||
		errors.Is(err, pagination.ErrInvalidCursor) {
		server.ErrorWithCodeAndMessage(c, server.CodeBadRequest, err.Error())
		return
	}
	if errors.Is(err, audit.ErrNotFound) {
		server.ErrorWithCode(c, server.CodeNotFound)
		return
	}
	server.ErrorWithCodeAndMessage(c, server.CodeInternalServerError, err.Error())
}
