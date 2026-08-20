package index_test

import (
	"os"
	"strings"
	"testing"

	"github.com/achedges/financial-storage-go"
	"github.com/achedges/financial-storage-go/index"
	"github.com/achedges/go-assertions"
)

// DateIndex tests

func TestDateIndex_NewDateIndex(t *testing.T) {
	di := index.NewDateIndex()
	assertions.EqualInts(0, di.Size(), t)

	di.Add(20260101, 1000, 200)
	assertions.True(di.Contains(20260101), t)
	assertions.EqualInts(1, di.Size(), t)
}

func TestDateIndex_IsDirty(t *testing.T) {
	di := index.NewDateIndex()
	assertions.False(di.IsDirty(), t)
	di.SetDirty()
	assertions.True(di.IsDirty(), t)
	di.SetClean()
	assertions.False(di.IsDirty(), t)
}

func TestDateIndex_Next(t *testing.T) {
	di := index.NewDateIndex()
	di.Add(20260101, 0, 100)
	di.Add(20260102, 100, 100)

	nextnode := di.Next(20260101) // should be 20260102
	if nextnode == nil {
		t.Error("Expected to find next node, was nil")
		return
	}
	assertions.EqualUints(20260102, nextnode.GetDate(), t)
	assertions.EqualInts(100, nextnode.GetOffset(), t)
	assertions.EqualUints(100, nextnode.GetCount(), t)

	prevnode := di.Prev(20260102) // should be 20260101
	if prevnode == nil {
		t.Error("Expected to find prev node, was nil")
		return
	}
	assertions.EqualUints(20260101, prevnode.GetDate(), t)
	assertions.EqualInts(0, prevnode.GetOffset(), t)
	assertions.EqualUints(100, prevnode.GetCount(), t)

	nextnode = di.Next(20260102)
	assertions.True(nextnode == nil, t)

	prevnode = di.Prev(20260101)
	assertions.True(prevnode == nil, t)
}

// SymbolIndex tests

func TestSymbolIndex_NewSymbolIndex(t *testing.T) {
	adapter := TestAdapter[TestFileStoreItem]{}
	idx := index.NewSymbolIndex[TestFileStoreItem](&adapter)
	assertions.EqualUints(0, idx.Size(), t)
}

func TestSymbolIndex_Add(t *testing.T) {
	adapter := TestAdapter[storage.FileStoreItem]{}
	idx := index.NewSymbolIndex[storage.FileStoreItem](&adapter)
	idx.Add("TEST", index.NewNode(20260101, 0, 100))
	assertions.True(idx.Contains("TEST"), t)
	assertions.True(idx.Lookup("TEST", 20260101) != nil, t)
	assertions.True(idx.IsDirty("TEST"), t)
}

func TestSymbolIndex_Load(t *testing.T) {
	adapter := TestAdapter[storage.FileStoreItem]{}
	idx := index.NewSymbolIndex[storage.FileStoreItem](&adapter)
	idx.Load("TEST_LOAD")
	assertions.EqualUints(1, idx.Size(), t)
	assertions.True(idx.Contains("TEST_LOAD"), t)

	first := idx.Lookup("TEST_LOAD", 20260101)
	second := idx.Lookup("TEST_LOAD", 20260102)
	third := idx.Lookup("TEST_LOAD", 20260103)

	if first == nil || second == nil || third == nil {
		t.Error("Date lookup failed")
		t.FailNow()
	}

	assertions.EqualUints(20260101, first.GetDate(), t)
	assertions.EqualInts(0, first.GetOffset(), t)
	assertions.EqualUints(100, first.GetCount(), t)

	assertions.EqualUints(20260102, second.GetDate(), t)
	assertions.EqualInts(100, second.GetOffset(), t)
	assertions.EqualUints(100, second.GetCount(), t)

	assertions.EqualUints(20260103, third.GetDate(), t)
	assertions.EqualInts(200, third.GetOffset(), t)
	assertions.EqualUints(100, third.GetCount(), t)

}

func TestSymbolIndex_Persist(t *testing.T) {
	adapter := TestAdapter[storage.FileStoreItem]{}

	// delete the test file (if it exists)
	_ = os.Remove(adapter.GetIndexFilePath("TEST_PERSIST"))

	idx := index.NewSymbolIndex[storage.FileStoreItem](&adapter)
	idx.Load("TEST_PERSIST")

	idx.Add("TEST_PERSIST", index.NewNode(20260101, 0, 100))
	idx.Add("TEST_PERSIST", index.NewNode(20260102, 100, 100))
	idx.Add("TEST_PERSIST", index.NewNode(20260103, 200, 100))
	assertions.EqualUints(1, idx.Size(), t) // one symbol index

	idx.Persist()

	// check resulting file
	text, _ := os.ReadFile(adapter.GetIndexFilePath("TEST_PERSIST"))
	lines := strings.Split(strings.TrimSpace(string(text)), "\n")
	assertions.EqualInts(3, len(lines), t)
	assertions.EqualStrings("20260102,100,100", lines[0], t)
	assertions.EqualStrings("20260101,0,100", lines[1], t)
	assertions.EqualStrings("20260103,200,100", lines[2], t)

	_ = os.Remove(adapter.GetIndexFilePath("TEST_PERSIST"))
}

func TestSymbolIndex_First_Last(t *testing.T) {
	adapter := TestAdapter[storage.FileStoreItem]{}
	idx := index.NewSymbolIndex(&adapter)
	idx.Load("TEST_LOAD")

	firstnode := idx.First("TEST_LOAD")
	lastnode := idx.Last("TEST_LOAD")

	if firstnode == nil {
		t.Error("Unable to fetch first node")
		t.FailNow()
	}
	if lastnode == nil {
		t.Error("Unable to fetch last node")
		t.FailNow()
	}

	assertions.EqualUints(20260101, firstnode.GetDate(), t)
	assertions.EqualInts(0, firstnode.GetOffset(), t)
	assertions.EqualUints(100, firstnode.GetCount(), t)

	assertions.EqualUints(20260103, lastnode.GetDate(), t)
	assertions.EqualInts(200, lastnode.GetOffset(), t)
	assertions.EqualUints(100, lastnode.GetCount(), t)
}

func TestSymbolIndex_Prev_Next(t *testing.T) {
	adapter := TestAdapter[storage.FileStoreItem]{}
	idx := index.NewSymbolIndex[storage.FileStoreItem](&adapter)
	idx.Load("TEST_LOAD")

	prevnode := idx.Prev("TEST_LOAD", 20260102)
	if prevnode == nil {
		t.Error("Unable to fetch previous node")
		t.FailNow()
	}
	assertions.EqualUints(20260101, prevnode.GetDate(), t)

	nextnode := idx.Next("TEST_LOAD", 20260102)
	if nextnode == nil {
		t.Error("Unable to fetch next node")
		t.FailNow()
	}
	assertions.EqualUints(20260103, nextnode.GetDate(), t)

	prevnode = idx.Prev("TEST_LOAD", prevnode.GetDate())
	assertions.True(prevnode == nil, t)

	nextnode = idx.Next("TEST_LOAD", nextnode.GetDate())
	assertions.True(nextnode == nil, t)
}
