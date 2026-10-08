package pagination

func Wrap[T CursorAware](fetched []T, resolved *Resolved, cfg QueryConfig, totalCount int64) *Connection[T] {
	if resolved == nil {
		resolved = &Resolved{Direction: Forward, Limit: DefaultLimit}
	}

	hasMore := len(fetched) > resolved.Limit
	if hasMore {
		fetched = fetched[:resolved.Limit]
	}
	if resolved.Direction == Backward {
		for i, j := 0, len(fetched)-1; i < j; i, j = i+1, j-1 {
			fetched[i], fetched[j] = fetched[j], fetched[i]
		}
	}

	conn := &Connection[T]{
		TotalCount: totalCount,
		Edges:      make([]*Edge[T], 0, len(fetched)),
		PageInfo: &PageInfo{
			HasPreviousPage: false,
			HasNextPage:     false,
		},
	}
	if len(fetched) == 0 {
		return conn
	}

	for _, item := range fetched {
		cursor := EncodeCursor(item.GetCursorValue(cfg.SortField), item.GetID())
		conn.Edges = append(conn.Edges, &Edge[T]{Cursor: cursor, Node: item})
	}

	conn.PageInfo.StartCursor = conn.Edges[0].Cursor
	conn.PageInfo.EndCursor = conn.Edges[len(conn.Edges)-1].Cursor
	hasCursor := resolved.Cursor != nil && *resolved.Cursor != ""
	if resolved.Direction == Forward {
		conn.PageInfo.HasNextPage = hasMore
		conn.PageInfo.HasPreviousPage = hasCursor
	} else {
		conn.PageInfo.HasPreviousPage = hasMore
		conn.PageInfo.HasNextPage = hasCursor
	}
	return conn
}
