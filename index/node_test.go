package index_test

import (
	"testing"

	"github.com/achedges/financial-storage-go/index"
	"github.com/achedges/go-assertions"
)

func TestNode_NewNodeFromCsv(t *testing.T) {
	node := index.NewNodeFromCsv("1,2,3")
	assertions.EqualUints(uint64(1), node.GetDate(), t)
	assertions.EqualInts(int64(2), node.GetOffset(), t)
	assertions.EqualUints(uint32(3), node.GetCount(), t)
}

func TestNode_ToCsv(t *testing.T) {
	node := index.NewNodeFromCsv("1,2,3")
	assertions.EqualStrings("1,2,3", node.ToCsv(), t)
}

func TestNode_IsAfter_IsBefore(t *testing.T) {
	node1 := index.NewNodeFromCsv("20260101,10000,200")
	node2 := index.NewNodeFromCsv("20260102,10000,200")
	assertions.False(node1.IsAfter(node2), t)
	assertions.True(node1.IsBefore(node2), t)
	assertions.False(node2.IsBefore(node1), t)
	assertions.True(node2.IsAfter(node1), t)
}

func TestNode_UpdateOffset(t *testing.T) {
	node := index.NewNodeFromCsv("20260101,10000,200")

	// increase offset
	node.UpdateOffset(100)
	assertions.EqualInts(10100, node.GetOffset(), t)

	// decrease offset
	node.UpdateOffset(-100)
	assertions.EqualInts(10000, node.GetOffset(), t)

	// check 0 offset minimum
	node.UpdateOffset(-10001)
	assertions.EqualInts(0, node.GetOffset(), t)
}
