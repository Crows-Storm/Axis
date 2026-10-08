package audit

import (
	"sort"
	"strings"
)

const MaxGraphEvents = 500

type EventNode struct {
	Event    *AuditEvent  `json:"event"`
	Children []*EventNode `json:"children"`
}

type EventGraph struct {
	EventID string       `json:"event_id"`
	Roots   []*EventNode `json:"roots"`
}

func Forest(focus *AuditEvent, siblings []*AuditEvent) *EventGraph {
	if focus == nil {
		return &EventGraph{Roots: []*EventNode{}}
	}
	graph := &EventGraph{EventID: focus.EventID, Roots: []*EventNode{}}
	if strings.TrimSpace(focus.Metadata.CorrelationId) == "" {
		graph.Roots = []*EventNode{{Event: focus}}
		return graph
	}

	byID := make(map[string]*AuditEvent, len(siblings)+1)
	byID[focus.EventID] = focus
	for _, ev := range siblings {
		if ev == nil || ev.EventID == "" {
			continue
		}
		byID[ev.EventID] = ev
	}

	children := make(map[string][]*AuditEvent)
	for _, ev := range byID {
		parent := strings.TrimSpace(ev.Metadata.CausationId)
		if parent == "" || byID[parent] == nil {
			continue
		}
		children[parent] = append(children[parent], ev)
	}

	visited := make(map[string]struct{}, len(byID))
	roots := make([]*AuditEvent, 0)
	for _, ev := range byID {
		parent := strings.TrimSpace(ev.Metadata.CausationId)
		if parent == "" || byID[parent] == nil {
			roots = append(roots, ev)
		}
	}
	sortEvents(roots)
	for _, root := range roots {
		graph.Roots = append(graph.Roots, buildNode(root, children, visited))
	}
	return graph
}

func buildNode(ev *AuditEvent, children map[string][]*AuditEvent, visited map[string]struct{}) *EventNode {
	node := &EventNode{Event: ev, Children: []*EventNode{}}
	if ev == nil {
		return node
	}
	if _, seen := visited[ev.EventID]; seen {
		return node
	}
	visited[ev.EventID] = struct{}{}
	kids := append([]*AuditEvent(nil), children[ev.EventID]...)
	sortEvents(kids)
	for _, child := range kids {
		node.Children = append(node.Children, buildNode(child, children, visited))
	}
	return node
}

func sortEvents(events []*AuditEvent) {
	sort.Slice(events, func(i, j int) bool {
		if events[i] == nil {
			return false
		}
		if events[j] == nil {
			return true
		}
		if !events[i].Timestamp.Equal(events[j].Timestamp) {
			return events[i].Timestamp.Before(events[j].Timestamp)
		}
		return events[i].EventID < events[j].EventID
	})
}
