package menu

import (
	"fmt"
	"strconv"
	"strings"
)

type TreePath string

func NewTreePath(path string) (TreePath, error) {
	tp := TreePath(path)
	if err := tp.Validate(); err != nil {
		return "", err
	}
	return tp, nil
}

// BuildTreePath 根据父路径和当前ID构建路径
func BuildTreePath(parentPath TreePath, currentID int64) TreePath {
	if parentPath == "" || parentPath == "/" {
		return TreePath(fmt.Sprintf("/%d/", currentID))
	}
	parent := string(parentPath)
	if !strings.HasSuffix(parent, "/") {
		parent += "/"
	}
	return TreePath(fmt.Sprintf("%s%d/", parent, currentID))
}

func (t TreePath) Validate() error {
	if t == "" {
		return nil // 顶级节点可以为空
	}
	s := string(t)
	if !strings.HasPrefix(s, "/") || !strings.HasSuffix(s, "/") {
		return fmt.Errorf("invalid tree path format: %s, must start and end with '/'", s)
	}
	return nil
}

// AncestorIDs 提取所有祖先节点ID（不含自身）
func (t TreePath) AncestorIDs() []int64 {
	if t == "" {
		return nil
	}
	parts := strings.Split(strings.Trim(string(t), "/"), "/")
	if len(parts) <= 1 {
		return nil // 只有自身，没有祖先
	}
	// 排除最后一个（自身）
	ids := make([]int64, 0, len(parts)-1)
	for _, p := range parts[:len(parts)-1] {
		if id, err := strconv.ParseInt(p, 10, 64); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

// Contains 判断路径是否包含指定节点ID
func (t TreePath) Contains(nodeID int64) bool {
	return strings.Contains(string(t), fmt.Sprintf("/%d/", nodeID))
}

// Depth 返回路径深度
func (t TreePath) Depth() int {
	if t == "" {
		return 0
	}
	return strings.Count(string(t), "/") - 1
}

func (t TreePath) String() string {
	return string(t)
}
