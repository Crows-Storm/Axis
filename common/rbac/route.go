package rbac

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Requirement is what a request needs. Self means resource-owner APIs (current user from context).
type Requirement struct {
	Self   bool   `json:"self,omitempty"`
	Menu   string `json:"menu,omitempty"`
	Action string `json:"action,omitempty"`
}

type Route struct {
	Audience Audience
	Method   string
	Path     string
	Req      Requirement
}

func Self(audience Audience, method, path string) Route {
	return Route{Audience: audience, Method: method, Path: path, Req: Requirement{Self: true}}
}

func Need(audience Audience, method, path, menu, action string) Route {
	return Route{
		Audience: audience,
		Method:   method,
		Path:     path,
		Req:      Requirement{Menu: menu, Action: action},
	}
}

func (r Requirement) encode() (string, error) {
	if r.Self {
		return `{"self":true}`, nil
	}
	if strings.TrimSpace(r.Menu) == "" || strings.TrimSpace(r.Action) == "" {
		return "", fmt.Errorf("rbac route requires menu and action")
	}
	if _, ok := ParseAction(r.Action); !ok {
		return "", fmt.Errorf("unknown action %q", r.Action)
	}
	b, err := json.Marshal(Requirement{Menu: r.Menu, Action: strings.ToUpper(r.Action)})
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func decodeRequirement(raw string) (Requirement, error) {
	var req Requirement
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		return Requirement{}, fmt.Errorf("decode route: %w", err)
	}
	if req.Self {
		return req, nil
	}
	if req.Menu == "" || req.Action == "" {
		return Requirement{}, fmt.Errorf("route requirement missing menu or action")
	}
	if _, ok := ParseAction(req.Action); !ok {
		return Requirement{}, fmt.Errorf("unknown action %q", req.Action)
	}
	return req, nil
}
