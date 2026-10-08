package adapters

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/Crows-Storm/Axis/common/pagination"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
)

func (m UserModel) GetCursorValue(string) string {
	return strconv.FormatUint(m.ID, 10)
}

func (m UserModel) GetID() string {
	return strconv.FormatUint(m.ID, 10)
}

func userPageConfig() pagination.QueryConfig {
	cfg := pagination.NewestFirst("id", "id")
	cfg.Count = true
	return cfg
}

func toUserConnection(src *pagination.Connection[UserModel]) *pagination.Connection[*domain.User] {
	out := &pagination.Connection[*domain.User]{
		TotalCount: src.TotalCount,
		PageInfo:   src.PageInfo,
		Edges:      make([]*pagination.Edge[*domain.User], 0, len(src.Edges)),
	}
	for _, edge := range src.Edges {
		if edge == nil {
			continue
		}
		node := edge.Node
		out.Edges = append(out.Edges, &pagination.Edge[*domain.User]{
			Cursor: edge.Cursor,
			Node:   node.toDomain(),
		})
	}
	return out
}

func safePrefix(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range raw {
		if r == '%' || r == '_' || r == '\\' {
			continue
		}
		if unicode.IsSpace(r) {
			continue
		}
		b.WriteRune(r)
		if b.Len() >= 64 {
			break
		}
	}
	return b.String()
}
