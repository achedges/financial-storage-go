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
	return int(di.root.Size)
}

func (di *DateIndex) Add(date uint64, offset int64, count uint32) {
	di.root.AddItem(date, NewNode(date, offset, count))
}

func (di *DateIndex) Contains(date uint64) bool {
	return di.root.Contains(date)
}

func (di *DateIndex) Next(date uint64) *Node {
	if !di.root.Contains(date) {
		return nil
	}
	next := di.root.Next(di.root.Find(date))
	if next == nil {
		return nil
	}
	return next.GetValue()
}

func (di *DateIndex) Prev(date uint64) *Node {
	if !di.root.Contains(date) {
		return nil
	}
	prev := di.root.Prev(di.root.Find(date))
	if prev == nil {
		return nil
	}
	return prev.GetValue()
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
	return si.root.Size
}

func (si *SymbolIndex[T]) Contains(symbol string) bool {
	return si.root.Contains(symbol)
}

func (si *SymbolIndex[T]) Add(symbol string, node *Node) {
	if !si.Contains(symbol) {
		si.root.AddItem(symbol, NewDateIndex())
	}
	idx := si.root.Find(symbol).GetValue()
	idx.Add(node.GetDate(), node.GetOffset(), node.GetCount())
	idx.SetDirty()
}

func (si *SymbolIndex[T]) SetDirty(symbol string) {
	if !si.Contains(symbol) {
		return
	}
	idx := si.root.Find(symbol).GetValue()
	idx.SetDirty()
}

func (si *SymbolIndex[T]) IsDirty(symbol string) bool {
	if !si.Contains(symbol) {
		return false
	}
	idx := si.root.Find(symbol).GetValue()
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
		dateindex.root.AddItem(node.date, node)
	}

	si.root.AddItem(symbol, dateindex)
}

func (si *SymbolIndex[T]) Persist() {
	idx := si.root.Min()
	for idx != nil {
		if idx.GetValue().IsDirty() {
			dateindex := idx.GetValue()
			filepath := si.adapter.GetIndexFilePath(idx.GetKey())
			file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0644)
			if file == nil || err != nil {
				fmt.Printf("Couldn't write symbol index file for %s, aborting\n", idx.GetKey())
				return
			}

			for _, date := range dateindex.root.GetKeys(gotrees.TreeWalkBFS) {
				datenode := dateindex.root.Find(date).GetValue()
				if datenode.GetCount() > 0 {
					_, _ = file.WriteString(fmt.Sprintf("%s\n", datenode.ToCsv()))
				}
			}

			idx.GetValue().SetClean()
			_ = file.Close()
		}

		idx = si.root.Next(idx)
	}
}

func (si *SymbolIndex[T]) Lookup(symbol string, date uint64) *Node {
	si.Load(symbol)
	idx := si.root.Find(symbol)
	if idx == nil || !idx.GetValue().root.Contains(date) {
		return nil
	}

	return idx.GetValue().root.Find(date).GetValue()
}

func (si *SymbolIndex[T]) First(symbol string) *Node {
	si.Load(symbol)
	idx := si.root.Find(symbol)
	if idx == nil || idx.GetValue().root.Size == 0 {
		return nil
	}

	return idx.GetValue().root.Min().GetValue()
}

func (si *SymbolIndex[T]) Last(symbol string) *Node {
	si.Load(symbol)
	idx := si.root.Find(symbol)
	if idx == nil || idx.GetValue().root.Size == 0 {
		return nil
	}

	return idx.GetValue().root.Max().GetValue()
}

func (si *SymbolIndex[T]) Prev(symbol string, date uint64) *Node {
	si.Load(symbol)
	idx := si.root.Find(symbol)
	if idx == nil || !idx.GetValue().root.Contains(date) {
		return nil
	}
	return idx.GetValue().Prev(date)
}

func (si *SymbolIndex[T]) Next(symbol string, date uint64) *Node {
	si.Load(symbol)
	idx := si.root.Find(symbol)
	if idx == nil || !idx.GetValue().root.Contains(date) {
		return nil
	}
	return idx.GetValue().Next(date)
}
