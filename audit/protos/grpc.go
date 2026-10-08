package protos

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Crows-Storm/Axis/audit/app"
	"github.com/Crows-Storm/Axis/audit/app/query"
	"github.com/Crows-Storm/Axis/audit/domain/audit"
	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/genproto/auditpb"
	"github.com/Crows-Storm/Axis/common/pagination"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	auditpb.UnimplementedAuditServiceServer
	app app.Application
}

func NewGRPCServer(application app.Application) *GRPCServer {
	return &GRPCServer{app: application}
}

func (s *GRPCServer) PageRootEvents(ctx context.Context, request *auditpb.PageRootEventsRequest) (*auditpb.AuditEventConnection, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	q := query.PageRootsQuery{
		Args:      pageArgs(request),
		Service:   request.Service,
		EventName: request.EventName,
	}
	from, err := optionalTime(request.From)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "from must be RFC3339")
	}
	to, err := optionalTime(request.To)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "to must be RFC3339")
	}
	q.From = from
	q.To = to

	conn, err := s.app.Queries.PageRoots.Handle(ctx, q)
	if err != nil {
		return nil, statusFrom(err)
	}
	return toProtoConnection(conn), nil
}

func (s *GRPCServer) GetEventGraph(ctx context.Context, request *auditpb.GetEventGraphRequest) (*auditpb.AuditEventGraph, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	graph, err := s.app.Queries.GetEventGraph.Handle(ctx, query.GetEventGraphQuery{EventID: request.EventId})
	if err != nil {
		return nil, statusFrom(err)
	}
	return toProtoGraph(graph), nil
}

func pageArgs(request *auditpb.PageRootEventsRequest) pagination.Args {
	args := pagination.Args{}
	if request.First != nil {
		n := int(*request.First)
		args.First = &n
	}
	if request.After != nil {
		args.After = request.After
	}
	if request.Last != nil {
		n := int(*request.Last)
		args.Last = &n
	}
	if request.Before != nil {
		args.Before = request.Before
	}
	return args
}

func optionalTime(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	t, err := parseTimestamp(raw)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func toProtoConnection(conn *pagination.Connection[*audit.AuditEvent]) *auditpb.AuditEventConnection {
	out := &auditpb.AuditEventConnection{Edges: []*auditpb.AuditEventEdge{}}
	if conn == nil {
		return out
	}
	out.TotalCount = conn.TotalCount
	for _, edge := range conn.Edges {
		if edge == nil {
			continue
		}
		out.Edges = append(out.Edges, &auditpb.AuditEventEdge{
			Cursor: edge.Cursor,
			Node:   toProto(edge.Node),
		})
	}
	if conn.PageInfo != nil {
		out.PageInfo = &auditpb.PageInfo{
			StartCursor:     conn.PageInfo.StartCursor,
			EndCursor:       conn.PageInfo.EndCursor,
			HasPreviousPage: conn.PageInfo.HasPreviousPage,
			HasNextPage:     conn.PageInfo.HasNextPage,
		}
	}
	return out
}

func toProtoGraph(graph *audit.EventGraph) *auditpb.AuditEventGraph {
	out := &auditpb.AuditEventGraph{Roots: []*auditpb.AuditEventNode{}}
	if graph == nil {
		return out
	}
	out.EventId = graph.EventID
	for _, root := range graph.Roots {
		out.Roots = append(out.Roots, toProtoNode(root))
	}
	return out
}

func toProtoNode(node *audit.EventNode) *auditpb.AuditEventNode {
	if node == nil {
		return nil
	}
	out := &auditpb.AuditEventNode{Event: toProto(node.Event), Children: []*auditpb.AuditEventNode{}}
	for _, child := range node.Children {
		out.Children = append(out.Children, toProtoNode(child))
	}
	return out
}

func toProto(row *audit.AuditEvent) *auditpb.AuditEvent {
	if row == nil {
		return nil
	}
	return &auditpb.AuditEvent{
		Id:         row.ID,
		EventId:    row.EventID,
		EventName:  row.EventName,
		OperatorId: row.OperatorID,
		Category:   row.Category,
		Timestamp:  row.Timestamp.UTC().Format(time.RFC3339Nano),
		Metadata: &auditpb.Metadata{
			Service:       row.Metadata.Service,
			AggregateId:   row.Metadata.AggregateID,
			CorrelationId: row.Metadata.CorrelationId,
			CausationId:   row.Metadata.CausationId,
			Payload:       row.Metadata.Payload,
		},
	}
}

func parseTimestamp(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, errors.New("empty timestamp")
	}
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, raw)
}

func statusFrom(err error) error {
	if err == nil {
		return nil
	}
	var executed decorator.CommandExecutedError
	if errors.As(err, &executed) ||
		errors.Is(err, audit.ErrInvalidRecord) ||
		errors.Is(err, pagination.ErrInvalidArgs) ||
		errors.Is(err, pagination.ErrFirstLastConflict) ||
		errors.Is(err, pagination.ErrInvalidCursor) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	if errors.Is(err, audit.ErrNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}
