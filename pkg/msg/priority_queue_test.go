package msg

import (
	"testing"
)

func TestNewPriorityQueue(t *testing.T) {
	q := NewPriorityQueue(func() map[string]uint8 {
		return map[string]uint8{
			"a": 1,
			"b": 2,
			"c": 3,
		}
	})
	q.Write(&Metadata{
		TraceID:     "666",
		ServiceName: "a",
	})
	q.Write(&Metadata{
		TraceID:     "666",
		ServiceName: "b",
	})
	q.Write(&Metadata{
		TraceID:     "666",
		ServiceName: "c",
	})
	t.Log(<-q.Read())
}
