package storage

import (
	"os"
)

type FileStoreItem interface {
	GetDate() uint64
	GetId() uint64
	CompareTo(other FileStoreItem) int
}

type DataAdapter[T FileStoreItem] interface {
	GetDataPath() string
	GetDataFilePath(symbol string) string
	GetIndexPath() string
	GetIndexFilePath(symbol string) string
	GetRecordSizeBytes(n uint32) uint32
	Buffer() []byte
	FillBuffer(item *T)
	Read(file os.File, symbol string) *T
}
