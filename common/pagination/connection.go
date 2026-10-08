package pagination

import "errors"

var (
	ErrInvalidArgs       = errors.New("invalid pagination arguments")
	ErrFirstLastConflict = errors.New("cannot use both first and last")
	ErrInvalidCursor     = errors.New("invalid cursor")
)

type Connection[T any] struct {
	TotalCount int64      `json:"total_count"`
	Edges      []*Edge[T] `json:"edges"`
	PageInfo   *PageInfo  `json:"page_info"`
}

type Edge[T any] struct {
	Cursor string `json:"cursor"`
	Node   T      `json:"node"`
}

type PageInfo struct {
	StartCursor     string `json:"start_cursor"`
	EndCursor       string `json:"end_cursor"`
	HasPreviousPage bool   `json:"has_previous_page"`
	HasNextPage     bool   `json:"has_next_page"`
}

type Args struct {
	First  *int    `json:"first"`
	After  *string `json:"after"`
	Last   *int    `json:"last"`
	Before *string `json:"before"`
}

type Direction int

const (
	Forward Direction = iota
	Backward
)

type Resolved struct {
	Direction Direction
	Limit     int
	Cursor    *string
}

const (
	DefaultLimit = 20
	MaxLimit     = 100
)
