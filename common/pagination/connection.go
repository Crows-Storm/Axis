package pagination

type Connection[T any] struct {
	TotalCount int64      `json:"totalCount"`
	Edges      []*Edge[T] `json:"edges"`
	PageInfo   *PageInfo  `json:"pageInfo"`
}

type Edge[T any] struct {
	Cursor string `json:"cursor"`
	Node   T      `json:"node"`
}

type PageInfo struct {
	StartCursor     string `json:"startCursor"`
	EndCursor       string `json:"endCursor"`
	HasPreviousPage bool   `json:"hasPreviousPage"`
	HasNextPage     bool   `json:"hasNextPage"`
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
