package collaboration

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestUpdateLogAppendAndGet(t *testing.T) {
	log := NewUpdateLog(10)

	entry1, err := log.Append("default", "p1", []byte("update-1"), true)
	if err != nil {
		t.Fatalf("unexpected error appending: %v", err)
	}
	if entry1.Seq != 1 {
		t.Errorf("expected seq 1, got %d", entry1.Seq)
	}

	entry2, err := log.Append("default", "p2", []byte("update-2"), false)
	if err != nil {
		t.Fatalf("unexpected error appending: %v", err)
	}
	if entry2.Seq != 2 {
		t.Errorf("expected seq 2, got %d", entry2.Seq)
	}

	if log.Count() != 2 {
		t.Errorf("expected count 2, got %d", log.Count())
	}
	if log.LatestSeq() != 2 {
		t.Errorf("expected latest seq 2, got %d", log.LatestSeq())
	}

	all := log.GetAll()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if string(all[0].Data) != "update-1" || string(all[1].Data) != "update-2" {
		t.Errorf("mismatched payload content: %v", all)
	}

	// Test GetSince
	since1 := log.GetSince(1)
	if len(since1) != 1 || since1[0].Seq != 2 {
		t.Fatalf("expected 1 entry with seq 2, got %v", since1)
	}

	since2 := log.GetSince(2)
	if len(since2) != 0 {
		t.Fatalf("expected 0 entries since 2, got %v", since2)
	}
}

func TestUpdateLogCapacityLimit(t *testing.T) {
	maxEntries := 3
	log := NewUpdateLog(maxEntries)

	for i := 0; i < maxEntries; i++ {
		_, err := log.Append("default", "p1", []byte(fmt.Sprintf("update-%d", i)), true)
		if err != nil {
			t.Fatalf("unexpected error at index %d: %v", i, err)
		}
	}

	// Next append should fail
	_, err := log.Append("default", "p1", []byte("overflow"), true)
	if err == nil {
		t.Fatal("expected capacity error, got nil")
	}
	if !errors.Is(err, ErrUpdateLogCapacityExceeded) {
		t.Errorf("expected ErrUpdateLogCapacityExceeded, got %v", err)
	}

	if log.Count() != maxEntries {
		t.Errorf("expected log count to remain at %d, got %d", maxEntries, log.Count())
	}
}

func TestUpdateLogDefensiveCopy(t *testing.T) {
	log := NewUpdateLog(5)
	src := []byte("original")

	_, err := log.Append("default", "p1", src, true)
	if err != nil {
		t.Fatalf("append failed: %v", err)
	}

	// Mutate source buffer
	src[0] = 'X'

	entries := log.GetAll()
	if string(entries[0].Data) != "original" {
		t.Errorf("log data was mutated by source buffer mutation: got %s", string(entries[0].Data))
	}

	// Mutate retrieved slice
	entries[0].Data[0] = 'Z'
	entriesAfter := log.GetAll()
	if string(entriesAfter[0].Data) != "original" {
		t.Errorf("log data was mutated by retrieved slice mutation: got %s", string(entriesAfter[0].Data))
	}
}

func TestUpdateLogConcurrentAppendsAndReads(t *testing.T) {
	log := NewUpdateLog(1000)
	var wg sync.WaitGroup

	numWriters := 10
	appendsPerWriter := 50

	// Concurrent writers
	for w := 0; w < numWriters; w++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for i := 0; i < appendsPerWriter; i++ {
				_, err := log.Append("default", fmt.Sprintf("p%d", writerID), []byte("payload"), true)
				if err != nil {
					t.Errorf("writer %d failed: %v", writerID, err)
					return
				}
			}
		}(w)
	}

	// Concurrent readers
	for r := 0; r < 5; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				_ = log.GetAll()
				_ = log.Count()
			}
		}()
	}

	wg.Wait()

	expectedTotal := numWriters * appendsPerWriter
	if log.Count() != expectedTotal {
		t.Errorf("expected %d entries, got %d", expectedTotal, log.Count())
	}
}
