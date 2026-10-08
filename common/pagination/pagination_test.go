package pagination

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type item struct {
	id string
	v  string
}

func (i item) GetCursorValue(string) string { return i.v }
func (i item) GetID() string                { return i.id }

func TestResolveDefaultsForward(t *testing.T) {
	got, err := Args{}.Resolve()
	require.NoError(t, err)
	assert.Equal(t, Forward, got.Direction)
	assert.Equal(t, DefaultLimit, got.Limit)
}

func TestResolveRejectsConflicts(t *testing.T) {
	first, last := 10, 10
	_, err := Args{First: &first, Last: &last}.Resolve()
	assert.ErrorIs(t, err, ErrFirstLastConflict)

	before := "cur_x"
	_, err = Args{First: &first, Before: &before}.Resolve()
	assert.ErrorIs(t, err, ErrInvalidArgs)

	zero := 0
	_, err = Args{First: &zero}.Resolve()
	assert.ErrorIs(t, err, ErrInvalidArgs)
}

func TestResolveDoesNotMutateArgs(t *testing.T) {
	args := Args{}
	_, err := args.Resolve()
	require.NoError(t, err)
	assert.Nil(t, args.First)
}

func TestCursorRoundTrip(t *testing.T) {
	raw := EncodeCursor("2026-10-04T00:00:00Z", "evt-1")
	v, id, err := DecodeCursor(raw)
	require.NoError(t, err)
	assert.Equal(t, "2026-10-04T00:00:00Z", v)
	assert.Equal(t, "evt-1", id)
}

func TestDecodeCursorRejectsGarbage(t *testing.T) {
	_, _, err := DecodeCursor("nope")
	assert.ErrorIs(t, err, ErrInvalidCursor)
}

func TestSeekFollowsSortDirection(t *testing.T) {
	after := EncodeCursor("t", "id-1")
	cfg := NewestFirst("timestamp", "event_id")
	resolved, err := Args{After: &after}.Resolve()
	require.NoError(t, err)

	seek, err := NewSeek(resolved, cfg)
	require.NoError(t, err)
	require.NotNil(t, seek)
	assert.False(t, seek.Greater)
	assert.False(t, cfg.QueryAscending(Forward))
	assert.True(t, cfg.QueryAscending(Backward))
}

func TestWrapForwardHasNext(t *testing.T) {
	cfg := NewestFirst("timestamp", "id")
	resolved := &Resolved{Direction: Forward, Limit: 2}
	fetched := []item{{"a", "3"}, {"b", "2"}, {"c", "1"}}
	conn := Wrap(fetched, resolved, cfg, 0)
	require.Len(t, conn.Edges, 2)
	assert.True(t, conn.PageInfo.HasNextPage)
	assert.False(t, conn.PageInfo.HasPreviousPage)
	assert.Equal(t, "a", conn.Edges[0].Node.id)
	assert.Equal(t, "b", conn.Edges[1].Node.id)
}

func TestWrapBackwardReverses(t *testing.T) {
	cfg := NewestFirst("timestamp", "id")
	before := EncodeCursor("t", "x")
	resolved := &Resolved{Direction: Backward, Limit: 2, Cursor: &before}
	fetched := []item{{"c", "1"}, {"b", "2"}, {"a", "3"}}
	conn := Wrap(fetched, resolved, cfg, 0)
	require.Len(t, conn.Edges, 2)
	assert.Equal(t, "b", conn.Edges[0].Node.id)
	assert.Equal(t, "c", conn.Edges[1].Node.id)
	assert.True(t, conn.PageInfo.HasPreviousPage)
	assert.True(t, conn.PageInfo.HasNextPage)
}

func TestMongoSeekFilterLt(t *testing.T) {
	got := MongoSeekFilter("timestamp", "event_id", &Seek{Value: "t", ID: "e", Greater: false}, "t")
	require.Len(t, got, 1)
	assert.Equal(t, "$or", got[0].Key)
	or, ok := got[0].Value.(bson.A)
	require.True(t, ok)
	require.Len(t, or, 2)
}
