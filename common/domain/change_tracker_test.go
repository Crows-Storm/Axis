package domain_test

import (
	"testing"
	"time"

	"github.com/Crows-Storm/Axis/common/domain"
	"github.com/stretchr/testify/assert"
)

func TestChangeTracker(t *testing.T) {
	t.Run("record single change", func(t *testing.T) {
		tracker := domain.NewChangeTracker()

		tracker.RecordChange("email", "old@example.com", "new@example.com")

		assert.True(t, tracker.HasChanges())
		changes := tracker.GetChanges()
		assert.Len(t, changes, 1)
		assert.Equal(t, "email", changes[0].FieldName)
		assert.Equal(t, "old@example.com", changes[0].OldValue)
		assert.Equal(t, "new@example.com", changes[0].NewValue)
	})

	t.Run("record multiple changes", func(t *testing.T) {
		tracker := domain.NewChangeTracker()

		tracker.RecordChange("email", "old@example.com", "new@example.com")
		tracker.RecordChange("name", "John", "Jane")
		tracker.RecordChange("age", 25, 26)

		assert.True(t, tracker.HasChanges())
		changes := tracker.GetChanges()
		assert.Len(t, changes, 3)
	})

	t.Run("ignore same values", func(t *testing.T) {
		tracker := domain.NewChangeTracker()

		tracker.RecordChange("email", "same@example.com", "same@example.com")

		assert.False(t, tracker.HasChanges())
		assert.Len(t, tracker.GetChanges(), 0)
	})

	t.Run("clear changes", func(t *testing.T) {
		tracker := domain.NewChangeTracker()

		tracker.RecordChange("email", "old@example.com", "new@example.com")
		assert.True(t, tracker.HasChanges())

		tracker.Clear()
		assert.False(t, tracker.HasChanges())
		assert.Len(t, tracker.GetChanges(), 0)
	})
}

func TestDiffTracker(t *testing.T) {
	t.Run("track string changes", func(t *testing.T) {
		tracker := domain.NewDiffTracker()

		tracker.TrackStringChange("email", "old@example.com", "new@example.com")
		tracker.TrackStringChange("name", "John", "Jane")

		assert.True(t, tracker.HasChanges())
		changes := tracker.GetChanges()
		assert.Len(t, changes, 2)
	})

	t.Run("track int64 changes", func(t *testing.T) {
		tracker := domain.NewDiffTracker()

		tracker.TrackInt64Change("user_id", int64(100), int64(200))

		assert.True(t, tracker.HasChanges())
		changes := tracker.GetChanges()
		assert.Len(t, changes, 1)
		assert.Equal(t, "user_id", changes[0].FieldName)
	})

	t.Run("track int8 changes", func(t *testing.T) {
		tracker := domain.NewDiffTracker()

		tracker.TrackInt8Change("status", int8(0), int8(1))

		assert.True(t, tracker.HasChanges())
		changes := tracker.GetChanges()
		assert.Len(t, changes, 1)
	})

	t.Run("track bool changes", func(t *testing.T) {
		tracker := domain.NewDiffTracker()

		tracker.TrackBoolChange("is_active", false, true)

		assert.True(t, tracker.HasChanges())
		changes := tracker.GetChanges()
		assert.Len(t, changes, 1)
		assert.Equal(t, false, changes[0].OldValue)
		assert.Equal(t, true, changes[0].NewValue)
	})

	t.Run("get change map", func(t *testing.T) {
		tracker := domain.NewDiffTracker()

		tracker.TrackStringChange("email", "old@example.com", "new@example.com")
		tracker.TrackStringChange("name", "John", "Jane")

		changeMap := tracker.GetChangeMap()
		assert.Len(t, changeMap, 2)
		assert.Contains(t, changeMap, "email")
		assert.Contains(t, changeMap, "name")
		assert.Equal(t, "new@example.com", changeMap["email"].NewValue)
	})

	t.Run("get changed fields", func(t *testing.T) {
		tracker := domain.NewDiffTracker()

		tracker.TrackStringChange("email", "old@example.com", "new@example.com")
		tracker.TrackStringChange("name", "John", "Jane")
		tracker.TrackInt8Change("status", int8(0), int8(1))

		fields := tracker.GetChangedFields()
		assert.Len(t, fields, 3)
		assert.Contains(t, fields, "email")
		assert.Contains(t, fields, "name")
		assert.Contains(t, fields, "status")
	})

	t.Run("ignore unchanged values", func(t *testing.T) {
		tracker := domain.NewDiffTracker()

		tracker.TrackStringChange("email", "same@example.com", "same@example.com")
		tracker.TrackInt64Change("id", int64(100), int64(100))
		tracker.TrackBoolChange("active", true, true)

		assert.False(t, tracker.HasChanges())
		assert.Len(t, tracker.GetChanges(), 0)
	})
}

func TestFieldChange(t *testing.T) {
	t.Run("field change string format", func(t *testing.T) {
		change := domain.FieldChange{
			FieldName: "email",
			OldValue:  "old@example.com",
			NewValue:  "new@example.com",
			Timestamp: time.Now(),
		}

		str := change.String()
		assert.Contains(t, str, "email")
		assert.Contains(t, str, "old@example.com")
		assert.Contains(t, str, "new@example.com")
	})

	t.Run("field change JSON marshaling", func(t *testing.T) {
		change := domain.FieldChange{
			FieldName: "status",
			OldValue:  int8(0),
			NewValue:  int8(1),
			Timestamp: time.Now(),
		}

		data, err := change.MarshalJSON()
		assert.NoError(t, err)
		assert.NotEmpty(t, data)
		assert.Contains(t, string(data), "status")
		assert.Contains(t, string(data), "timestamp")
	})
}
