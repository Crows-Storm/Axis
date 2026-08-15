package pagination

import (
	"fmt"

	"gorm.io/gorm"
)

// QueryConfig Pageination query config
type QueryConfig struct {
	SortField string
	SortOrder string
	IDField   string
}

// DefaultConfig Default the config
func DefaultConfig() QueryConfig {
	return QueryConfig{
		SortField: "CreatedTime",
		SortOrder: "ASC",
		IDField:   "Id",
	}
}

// FetchConnection 通用游标分页查询引擎
//
// principle：
//
//	Forward (first + after):
//	  WHERE (sortField > cursorV) OR (sortField = cursorV AND id > cursorID)
//	  ORDER BY sortField ASC, id ASC
//	  LIMIT (limit + 1)
//
//	Backward (last + before):
//	  WHERE (sortField < cursorV) OR (sortField = cursorV AND id < cursorID)
//	  ORDER BY sortField DESC, id DESC
//	  LIMIT (limit + 1)
func FetchConnection[T any](
	db *gorm.DB,
	model *gorm.DB,
	config QueryConfig,
	args *Args,
) (*Connection[T], error) {

	resolved, err := args.Resolve()
	if err != nil {
		return nil, err
	}

	var totalCount int64
	if err := model.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("count failed: %w", err)
	}

	limit := resolved.Limit + 1

	orderField := config.SortField
	orderDir := config.SortOrder // ASC
	idField := config.IDField

	if resolved.Direction == Backward {
		if orderDir == "ASC" {
			orderDir = "DESC"
		} else {
			orderDir = "ASC"
		}
	}

	// Building a query
	query := model.Session(&gorm.Session{NewDB: false})

	if resolved.Cursor != nil && *resolved.Cursor != "" {
		cursorV, cursorID, err := DecodeCursor(*resolved.Cursor)
		if err != nil {
			return nil, fmt.Errorf("decode cursor failed: %w", err)
		}

		if resolved.Direction == Forward {
			// WHERE (sort > v) OR (sort = v AND id > cursorID)
			query = query.Where(
				fmt.Sprintf("(%s > ?) OR (%s = ? AND %s > ?)", orderField, orderField, idField),
				cursorV, cursorV, cursorID,
			)
		} else {
			// Backward: WHERE (sort < v) OR (sort = v AND id < cursorID)
			query = query.Where(
				fmt.Sprintf("(%s < ?) OR (%s = ? AND %s < ?)", orderField, orderField, idField),
				cursorV, cursorV, cursorID,
			)
		}
	}

	query = query.
		Order(fmt.Sprintf("%s %s, %s %s", orderField, orderDir, idField, orderDir)).
		Limit(limit)

	var results []T
	if err := query.Find(&results).Error; err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	hasMore := len(results) > resolved.Limit
	if hasMore {
		results = results[:resolved.Limit]
	}

	if resolved.Direction == Backward {
		for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
			results[i], results[j] = results[j], results[i]
		}
	}

	conn := buildConnection[T](results, config, totalCount, resolved, hasMore)
	return conn, nil
}

func buildConnection[T any](
	results []T,
	config QueryConfig,
	totalCount int64,
	resolved *Resolved,
	hasMore bool,
) *Connection[T] {

	conn := &Connection[T]{
		TotalCount: totalCount,
		Edges:      make([]*Edge[T], 0, len(results)),
		PageInfo: &PageInfo{
			HasPreviousPage: false,
			HasNextPage:     false,
		},
	}

	if len(results) == 0 {
		return conn
	}

	for _, item := range results {
		aware, ok := any(item).(CursorAware[any])
		if !ok {
			conn.Edges = append(conn.Edges, &Edge[T]{Node: item})
			continue
		}
		cursorValue := aware.GetCursorValue(config.SortField)
		cursorID := aware.GetID()
		cursor := EncodeCursor(cursorValue, cursorID)
		conn.Edges = append(conn.Edges, &Edge[T]{
			Cursor: cursor,
			Node:   item,
		})
	}

	// PageInfo
	conn.PageInfo.StartCursor = conn.Edges[0].Cursor
	conn.PageInfo.EndCursor = conn.Edges[len(conn.Edges)-1].Cursor

	if resolved.Direction == Forward {
		conn.PageInfo.HasNextPage = hasMore
		conn.PageInfo.HasPreviousPage = resolved.Cursor != nil && *resolved.Cursor != ""
	} else {
		conn.PageInfo.HasPreviousPage = hasMore
		conn.PageInfo.HasNextPage = resolved.Cursor != nil && *resolved.Cursor != ""
	}

	return conn
}
