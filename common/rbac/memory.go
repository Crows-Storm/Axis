package rbac

import (
	"context"
	"sync"
)

type Memory struct {
	mu     sync.RWMutex
	kv     map[string]string
	sets   map[string]map[string]struct{}
	hashes map[string]map[string]string
}

func NewMemory() *Memory {
	return &Memory{
		kv:     map[string]string{},
		sets:   map[string]map[string]struct{}{},
		hashes: map[string]map[string]string{},
	}
}

func (m *Memory) Get(_ context.Context, key string) (string, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.kv[key]
	return v, ok, nil
}

func (m *Memory) Set(_ context.Context, key, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.kv[key] = value
	return nil
}

func (m *Memory) Del(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.kv, key)
	delete(m.sets, key)
	delete(m.hashes, key)
	return nil
}

func (m *Memory) Exists(_ context.Context, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, ok := m.kv[key]; ok {
		return true, nil
	}
	if s, ok := m.sets[key]; ok && len(s) > 0 {
		return true, nil
	}
	if h, ok := m.hashes[key]; ok && len(h) > 0 {
		return true, nil
	}
	return false, nil
}

func (m *Memory) SMembers(_ context.Context, key string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	set := m.sets[key]
	out := make([]string, 0, len(set))
	for member := range set {
		out = append(out, member)
	}
	return out, nil
}

func (m *Memory) SAdd(_ context.Context, key string, members ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	set := m.sets[key]
	if set == nil {
		set = map[string]struct{}{}
		m.sets[key] = set
	}
	for _, member := range members {
		set[member] = struct{}{}
	}
	return nil
}

func (m *Memory) HGet(_ context.Context, key, field string) (string, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	h := m.hashes[key]
	if h == nil {
		return "", false, nil
	}
	v, ok := h[field]
	return v, ok, nil
}

func (m *Memory) HSet(_ context.Context, key string, fields map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	h := m.hashes[key]
	if h == nil {
		h = map[string]string{}
		m.hashes[key] = h
	}
	for field, value := range fields {
		h[field] = value
	}
	return nil
}
