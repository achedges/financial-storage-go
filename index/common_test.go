package index_test

import (
	"fmt"
	"os"

	"github.com/achedges/financial-storage-go"
)

type TestFileStoreItem struct {
	date uint64
	id   uint64
}

func (t TestFileStoreItem) CompareTo(other storage.FileStoreItem) int {
	if t.id < other.Id() {
		return -1
	} else if t.id > other.Id() {
		return 1
	}
	return 0
}

func (t TestFileStoreItem) Date() uint64 {
	return t.date
}

func (t TestFileStoreItem) Id() uint64 {
	return t.id
}

type TestAdapter[T storage.FileStoreItem] struct {
	item T
}

func (a *TestAdapter[T]) GetDataPath() string {
	return ""
}

func (a *TestAdapter[T]) GetDataFilePath(symbol string) string {
	return fmt.Sprintf("%s/%s", a.GetDataPath(), symbol)
}

func (a *TestAdapter[T]) GetIndexPath() string {
	return "."
}

func (a *TestAdapter[T]) GetIndexFilePath(symbol string) string {
	return fmt.Sprintf("%s/%s.csv", a.GetIndexPath(), symbol)
}

func (a *TestAdapter[T]) GetRecordSizeBytes(n uint32) uint32 {
	return 0
}

func (a *TestAdapter[T]) Buffer() []byte {
	return nil
}

func (a *TestAdapter[T]) FillBuffer(item *T) {
	return
}

func (a *TestAdapter[T]) Read(file os.File, symbol string) *T {
	return &(a.item)
}
