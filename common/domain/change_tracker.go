// Package domain provides change tracking capabilities for domain entities
package domain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

type ChangeTracker interface {
	RecordChange(fieldName string, oldValue, newValue interface{})
	GetChanges() []FieldChange
	HasChanges() bool
	Clear()
}

type FieldChange struct {
	FieldName string      `json:"field_name"`
	OldValue  interface{} `json:"old_value"`
	NewValue  interface{} `json:"new_value"`
	Timestamp time.Time   `json:"timestamp"`
}

func (fc FieldChange) String() string {
	return fmt.Sprintf("%s: %v -> %v", fc.FieldName, fc.OldValue, fc.NewValue)
}

func (fc FieldChange) MarshalJSON() ([]byte, error) {
	type Alias FieldChange
	return json.Marshal(&struct {
		Timestamp int64 `json:"timestamp"`
		*Alias
	}{
		Timestamp: fc.Timestamp.UnixMilli(),
		Alias:     (*Alias)(&fc),
	})
}

type changeTracker struct {
	changes []FieldChange
}

func NewChangeTracker() ChangeTracker {
	return &changeTracker{
		changes: make([]FieldChange, 0),
	}
}

func (ct *changeTracker) RecordChange(fieldName string, oldValue, newValue interface{}) {
	if reflect.DeepEqual(oldValue, newValue) {
		return
	}

	ct.changes = append(ct.changes, FieldChange{
		FieldName: fieldName,
		OldValue:  oldValue,
		NewValue:  newValue,
		Timestamp: time.Now(),
	})
}

func (ct *changeTracker) GetChanges() []FieldChange {
	return ct.changes
}

func (ct *changeTracker) HasChanges() bool {
	return len(ct.changes) > 0
}

func (ct *changeTracker) Clear() {
	ct.changes = make([]FieldChange, 0)
}

type DiffTracker struct {
	tracker ChangeTracker
}

func NewDiffTracker() *DiffTracker {
	return &DiffTracker{
		tracker: NewChangeTracker(),
	}
}

func (dt *DiffTracker) TrackStringChange(fieldName string, oldVal, newVal string) {
	if oldVal != newVal {
		dt.tracker.RecordChange(fieldName, oldVal, newVal)
	}
}

func (dt *DiffTracker) TrackInt64Change(fieldName string, oldVal, newVal int64) {
	if oldVal != newVal {
		dt.tracker.RecordChange(fieldName, oldVal, newVal)
	}
}

func (dt *DiffTracker) TrackInt8Change(fieldName string, oldVal, newVal int8) {
	if oldVal != newVal {
		dt.tracker.RecordChange(fieldName, oldVal, newVal)
	}
}

func (dt *DiffTracker) TrackBoolChange(fieldName string, oldVal, newVal bool) {
	if oldVal != newVal {
		dt.tracker.RecordChange(fieldName, oldVal, newVal)
	}
}

func (dt *DiffTracker) TrackChange(fieldName string, oldVal, newVal interface{}) {
	dt.tracker.RecordChange(fieldName, oldVal, newVal)
}

func (dt *DiffTracker) GetChanges() []FieldChange {
	return dt.tracker.GetChanges()
}

func (dt *DiffTracker) HasChanges() bool {
	return dt.tracker.HasChanges()
}

func (dt *DiffTracker) GetChangeMap() map[string]FieldChange {
	changes := dt.tracker.GetChanges()
	changeMap := make(map[string]FieldChange, len(changes))

	for _, change := range changes {
		changeMap[change.FieldName] = change
	}

	return changeMap
}

func (dt *DiffTracker) GetChangedFields() []string {
	changes := dt.tracker.GetChanges()
	fields := make([]string, 0, len(changes))

	for _, change := range changes {
		fields = append(fields, change.FieldName)
	}

	return fields
}
