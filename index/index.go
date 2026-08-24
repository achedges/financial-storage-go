package index

import (
	"fmt"
	"os"
	"strings"

	"github.com/achedges/financial-storage-go"
	"github.com/achedges/gotrees"
)

// DateIndex implementation

type DateIndex struct {
	root    *gotrees.TreeMap[uint64, *Node]
	isDirty bool
}

func NewDateIndex() *DateIndex {
	return &DateIndex{
		root:    gotrees.NewTreeMap[uint64, *Node](),
		isDirty: false,
	}
}

func (di *DateIndex) SetDirty() {
	di.isDirty = true
}

func (di *DateIndex) SetClean() {
	di.isDirty = false
}

func (di *DateIndex) IsDirty() bool {
	return di.isDirty
}

func (di *DateIndex) Size() int {
	return int(di.root.Size())
}

func (di *DateIndex) Add(date uint64, offset int64, count uint32) {
	di.root.Add(date, NewNode(date, offset, count))
}

func (di *DateIndex) Contains(date uint64) bool {
	return di.root.Contains(date)
}

func (di *DateIndex) Next(date uint64) *Node {
	if !di.root.Contains(date) {
		return nil
	}
	next, found := di.root.Next(date)
	if !found {
		return nil
	}
	node, _ := di.root.Get(next)
	return node
}

func (di *DateIndex) Prev(date uint64) *Node {
	if !di.root.Contains(date) {
		return nil
	}
	prev, found := di.root.Prev(date)
	if !found {
		return nil
	}
	node, _ := di.root.Get(prev)
	return node
}

// SymbolIndex implementation

type SymbolIndex[T storage.FileStoreItem] struct {
	root    *gotrees.TreeMap[string, *DateIndex]
	adapter storage.DataAdapter[T]
}

func NewSymbolIndex[T storage.FileStoreItem](adapter storage.DataAdapter[T]) *SymbolIndex[T] {
	return &SymbolIndex[T]{
		root:    gotrees.NewTreeMap[string, *DateIndex](),
		adapter: adapter,
	}
}

func (si *SymbolIndex[T]) Size() uint32 {
	return si.root.Size()
}

func (si *SymbolIndex[T]) Contains(symbol string) bool {
	return si.root.Contains(symbol)
}

func (si *SymbolIndex[T]) Add(symbol string, node *Node) {
	if !si.Contains(symbol) {
		si.root.Add(symbol, NewDateIndex())
	}
	idx, _ := si.root.Get(symbol)
	idx.Add(node.GetDate(), node.GetOffset(), node.GetCount())
	idx.SetDirty()
}

func (si *SymbolIndex[T]) SetDirty(symbol string) {
	if !si.Contains(symbol) {
		return
	}
	idx, _ := si.root.Get(symbol)
	idx.SetDirty()
}

func (si *SymbolIndex[T]) IsDirty(symbol string) bool {
	if !si.Contains(symbol) {
		return false
	}
	idx, _ := si.root.Get(symbol)
	return idx.IsDirty()
}

func (si *SymbolIndex[T]) Load(symbol string) {
	if si.root.Contains(symbol) {
		return
	}

	dateindex := NewDateIndex()
	filepath := si.adapter.GetIndexFilePath(symbol)
	text, err := os.ReadFile(filepath)
	if err != nil {
		return
	}

	lines := strings.Split(strings.TrimSpace(string(text)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		node := NewNodeFromCsv(strings.TrimSpace(line))
		if node == nil {
			return
		}
		dateindex.root.Add(node.date, node)
	}

	si.root.Add(symbol, dateindex)
}

func (si *SymbolIndex[T]) Persist() {
	symbol, hasSymbolIndex := si.root.Min()
	for hasSymbolIndex {
		dateindex, _ := si.root.Get(symbol)
		if dateindex.IsDirty() {
			filepath := si.adapter.GetIndexFilePath(symbol)
			file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0644)
			if file == nil || err != nil {
				fmt.Printf("Couldn't write symbol index file for %s, aborting\n", symbol)
				return
			}

			for _, date := range dateindex.root.GetKeys(gotrees.TreeWalkBFS) {
				datenode, _ := dateindex.root.Get(date)
				if datenode.GetCount() > 0 {
					_, _ = file.WriteString(fmt.Sprintf("%s\n", datenode.ToCsv()))
				}
			}

			dateindex.SetClean()
			_ = file.Close()
		}

		symbol, hasSymbolIndex = si.root.Next(symbol)
	}
}

func (si *SymbolIndex[T]) Lookup(symbol string, date uint64) *Node {
	si.Load(symbol)
	
	sindex, hasSymbol := si.root.Get(symbol)
	if !hasSymbol {
		return nil
	}

	dindex, hasDate := sindex.root.Get(date)
	if !hasDate {
		return nil
	}

	return dindex
}

func (si *SymbolIndex[T]) First(symbol string) *Node {
	si.Load(symbol)

	sindex, hasSymbol := si.root.Get(symbol)
	if !hasSymbol {
		return nil
	}

	minKey, hasMin := sindex.root.Min()
	if !hasMin {
		return nil
	}

	dindex, hasDate := sindex.root.Get(minKey)
	if !hasDate {
		return nil
	}

	return dindex
}

func (si *SymbolIndex[T]) Last(symbol string) *Node {
	si.Load(symbol)

	sindex, hasSymbol := si.root.Get(symbol)
	if !hasSymbol {
		return nil
	}

	maxKey, hasMax := sindex.root.Max()
	if !hasMax {
		return nil
	}

	dindex, hasDate := sindex.root.Get(maxKey)
	if !hasDate {
		return nil
	}

	return dindex
}

func (si *SymbolIndex[T]) Prev(symbol string, date uint64) *Node {
	si.Load(symbol)

	sindex, hasSymbol := si.root.Get(symbol)
	if !hasSymbol {
		return nil
	}

	prevKey, hasPrev := sindex.root.Prev(date)
	if !hasPrev {
		return nil
	}

	prevNode, _ := sindex.root.Get(prevKey)
	return prevNode
}

func (si *SymbolIndex[T]) Next(symbol string, date uint64) *Node {
	si.Load(symbol)

	sindex, hasSymbol := si.root.Get(symbol)
	if !hasSymbol {
		return nil
	}

	nextKey, hasNext := sindex.root.Next(date)
	if !hasNext {
		return nil
	}

	nextNode, _ := sindex.root.Get(nextKey)
	return nextNode
}
