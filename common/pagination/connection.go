package pagination

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
	First  *int    `json:"first"  validate:"omitempty,min=1,max=100"`
	After  *string `json:"after"`
	Last   *int    `json:"last"   validate:"omitempty,min=1,max=100"`
	Before *string `json:"before"`
}

type Direction int

const (
	Forward Direction = iota // 向后翻（first + after）
	Backward
)

type Resolved struct {
	Direction Direction
	Limit     int
	Cursor    *string // Decoded cursor value (Value of the sorting field)
	CursorID  *string // Decoded cursor value (ID used for deduplication)
}

const (
	DefaultLimit = 20
	MaxLimit     = 100
)
