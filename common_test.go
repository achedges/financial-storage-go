package filestore_test

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path"

	"github.com/achedges/financial-storage-core-go/adapter"
)

// test item implementation

type SimpleFileStoreItem struct {
	id    uint64
	date  uint64
	time  uint32
	price float64
}

func (s SimpleFileStoreItem) GetId() uint64 {
	return s.id
}

func (s SimpleFileStoreItem) GetDate() uint64 {
	return s.date
}

func (s SimpleFileStoreItem) CompareTo(other adapter.FileStoreItem) int {
	if s.id < other.GetId() {
		return -1
	} else if s.id > other.GetId() {
		return 1
	}
	return 0
}

func (s SimpleFileStoreItem) GenerateId() uint64 {
	return (s.date * 10000) + uint64(s.time)
}

type SimpleFileStoreItemAdapter[T SimpleFileStoreItem] struct {
	rootPath        string
	dataPath        string
	indexPath       string
	recordSizeBytes uint32
	buffer          []byte
}

func NewSimpleFileStoreItemAdapter[T SimpleFileStoreItem](rootPath string, dataPath string, indexPath string) *SimpleFileStoreItemAdapter[T] {
	a := SimpleFileStoreItemAdapter[T]{
		rootPath:  rootPath,
		dataPath:  dataPath,
		indexPath: indexPath,
	}

	// id is not persisted
	a.recordSizeBytes = 8 + 4 + 8
	a.buffer = make([]byte, a.recordSizeBytes)
	return &a
}

func (s *SimpleFileStoreItemAdapter[T]) GetDataPath() string {
	return path.Join(s.rootPath, s.dataPath)
}

func (s *SimpleFileStoreItemAdapter[T]) GetDataFilePath(symbol string) string {
	return path.Join(s.GetDataPath(), symbol)
}

func (s *SimpleFileStoreItemAdapter[T]) GetIndexPath() string {
	return path.Join(s.rootPath, s.indexPath)
}

func (s *SimpleFileStoreItemAdapter[T]) GetIndexFilePath(symbol string) string {
	return path.Join(s.GetIndexPath(), fmt.Sprintf("%s.csv", symbol))
}

func (s *SimpleFileStoreItemAdapter[T]) GetRecordSizeBytes(n uint32) uint32 {
	return s.recordSizeBytes * n
}

func (s *SimpleFileStoreItemAdapter[T]) Buffer() []byte {
	return s.buffer
}

func (s *SimpleFileStoreItemAdapter[T]) FillBuffer(item *SimpleFileStoreItem) {
	binary.LittleEndian.PutUint64(s.buffer, item.date)
	binary.LittleEndian.PutUint32(s.buffer[8:], item.time)
	binary.LittleEndian.PutUint64(s.buffer[12:], math.Float64bits(item.price))
}

func (s *SimpleFileStoreItemAdapter[T]) Read(file os.File, _ string) *SimpleFileStoreItem {
	bytesRead, _ := file.Read(s.buffer)
	if uint32(bytesRead) < s.GetRecordSizeBytes(1) {
		return nil
	}

	item := SimpleFileStoreItem{
		date:  binary.LittleEndian.Uint64(s.buffer),
		time:  binary.LittleEndian.Uint32(s.buffer[8:]),
		price: math.Float64frombits(binary.LittleEndian.Uint64(s.buffer[12:])),
	}
	item.id = item.GenerateId()
	return &item
}
