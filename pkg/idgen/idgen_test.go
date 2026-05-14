package idgen

import (
	"testing"
)

func TestSnowflake_NextID_Unique(t *testing.T) {
	n := 10000
	sf := New(1)
	ids := make(map[int64]bool, n)

	for range n {
		id := sf.NextID()
		if ids[id] {
			t.Fatalf("duplicate id: %d", id)
		}
		ids[id] = true
	}
}

func TestSnowflake_NextID_MonotonicallyIncreasing(t *testing.T) {
	n := 10000
	sf := New(1)
	var prev int64

	for i := range n {
		id := sf.NextID()
		if i > 0 && id <= prev {
			t.Fatalf("id %d <= previous %d", id, prev)
		}
		prev = id
	}
}

func TestSnowflake_NextID_DifferentWorkers(t *testing.T) {
	sf1 := New(1)
	sf2 := New(2)
	ids := make(map[int64]bool)

	for range 5000 {
		id1 := sf1.NextID()
		id2 := sf2.NextID()
		if ids[id1] {
			t.Fatalf("duplicate id from worker 1: %d", id1)
		}
		if ids[id2] {
			t.Fatalf("duplicate id from worker 2: %d", id2)
		}
		ids[id1] = true
		ids[id2] = true
	}
}
