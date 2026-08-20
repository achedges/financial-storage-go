package filestore

import (
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/achedges/financial-storage-go"
	"github.com/achedges/financial-storage-go/index"
)

type FileStore[T storage.FileStoreItem] struct {
	adapter storage.DataAdapter[T]
	index   *index.SymbolIndex[T]
}

func NewFileStore[T storage.FileStoreItem](adapter storage.DataAdapter[T]) *FileStore[T] {
	return &FileStore[T]{
		adapter: adapter,
		index:   index.NewSymbolIndex[T](adapter),
	}
}

func (fs *FileStore[T]) shiftBytes(file os.File, offset int64, count uint32, deltabytes int64) {
	localBuffer := make([]byte, fs.adapter.GetRecordSizeBytes(count))
	_, _ = file.ReadAt(localBuffer, offset)
	_, _ = file.WriteAt(localBuffer, offset+deltabytes)
}

func (fs *FileStore[T]) writeBytes(file os.File, items []T, offset int64) {
	_, _ = file.Seek(offset, 0)
	for _, item := range items {
		fs.adapter.FillBuffer(&item)
		_, _ = file.Write(fs.adapter.Buffer())
	}
}

func (fs *FileStore[T]) writeNewItems(symbol string, date uint64, items []T, file os.File) {
	newNode := index.NewNode(date, 0, uint32(len(items)))
	lastNode := fs.index.Last(symbol)

	if lastNode != nil {
		if newNode.IsAfter(lastNode) {
			// append to the end of the file
			newNode.UpdateOffset(int64(fs.adapter.GetRecordSizeBytes(lastNode.GetCount())))
		} else {
			//shift newer records and insert into the file
			deltaBytes := int64(fs.adapter.GetRecordSizeBytes(newNode.GetCount()))
			shiftNode := lastNode
			for shiftNode != nil && newNode.IsBefore(shiftNode) {
				fs.shiftBytes(file, shiftNode.GetOffset(), shiftNode.GetCount(), deltaBytes)
				shiftNode.UpdateOffset(deltaBytes)
				shiftNode = fs.index.Prev(symbol, shiftNode.GetDate())
			}

			if shiftNode != nil {
				newNode.SetOffset(shiftNode.GetOffset() + int64(fs.adapter.GetRecordSizeBytes(shiftNode.GetCount())))
			}
		}
	}

	fs.index.Add(symbol, newNode) // calls SetDirty()
	fs.writeBytes(file, items, newNode.GetOffset())
}

func (fs *FileStore[T]) writeExistingItems(symbol string, updateNode *index.Node, items []T, file os.File) {
	numItems := uint32(len(items))
	if updateNode.GetCount() < numItems {
		// new buffer has more data, need to expand the file
		shiftNode := fs.index.Last(symbol)
		deltaBytes := int64(fs.adapter.GetRecordSizeBytes(numItems - updateNode.GetCount()))

		for shiftNode != nil && shiftNode.IsAfter(updateNode) {
			fs.shiftBytes(file, shiftNode.GetOffset(), shiftNode.GetCount(), deltaBytes)
			shiftNode.UpdateOffset(deltaBytes)
			shiftNode = fs.index.Prev(symbol, shiftNode.GetDate())
		}

		updateNode.SetCount(numItems)
	} else if updateNode.GetCount() > numItems {
		// new buffer has less data, need to compact the file
		shiftNode := fs.index.Next(symbol, updateNode.GetDate())
		deltaBytes := int64(fs.adapter.GetRecordSizeBytes(updateNode.GetCount()-numItems)) * -1

		for shiftNode != nil {
			fs.shiftBytes(file, shiftNode.GetOffset(), shiftNode.GetCount(), deltaBytes)
			shiftNode.UpdateOffset(deltaBytes)
			shiftNode = fs.index.Next(symbol, shiftNode.GetDate())
		}

		// trim file
		lastNode := fs.index.Last(symbol)
		if lastNode != nil {
			_ = file.Truncate(lastNode.GetOffset() + int64(fs.adapter.GetRecordSizeBytes(lastNode.GetCount())))
		}

		updateNode.SetCount(numItems)
	}

	fs.index.SetDirty(symbol)
	fs.writeBytes(file, items, updateNode.GetOffset())
}

func (fs *FileStore[T]) Write(symbol string, date uint64, items []T) {
	fs.index.Load(symbol)

	// sort the items first
	slices.SortFunc(items, func(a, b T) int { return a.CompareTo(b) })
	fileName := fs.adapter.GetDataFilePath(symbol)
	writeNode := fs.index.Lookup(symbol, date)

	file, e := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0644)
	if e != nil {
		panic(e)
	}
	if writeNode == nil {
		fs.writeNewItems(symbol, date, items, *file)
	} else {
		fs.writeExistingItems(symbol, writeNode, items, *file)
	}
	_ = file.Close()

	fs.index.SetDirty(symbol)
	fs.index.Persist()
}

func (fs *FileStore[T]) Read(symbol string, fromDate uint64, throughDate uint64) []*T {
	fs.index.Load(symbol)
	items := make([]*T, 0)

	datenode := fs.index.Lookup(symbol, fromDate)
	if datenode == nil {
		return items
	}

	filename := fs.adapter.GetDataFilePath(symbol)
	file, _ := os.OpenFile(filename, os.O_RDONLY, 0)
	for datenode.GetDate() <= throughDate {
		_, _ = file.Seek(datenode.GetOffset(), 0)
		for range datenode.GetCount() {
			item := fs.adapter.Read(*file, symbol)
			if item != nil {
				items = append(items, item)
			}
		}

		datenode = fs.index.Next(symbol, datenode.GetDate())
		if datenode == nil {
			break
		}
	}

	_ = file.Close()

	return items
}

func (fs *FileStore[T]) CheckIntegrity(symbol string) error {
	filename := fs.adapter.GetDataFilePath(symbol)
	file, _ := os.OpenFile(filename, os.O_RDONLY, 0)

	// read first record then rewind
	item := fs.adapter.Read(*file, symbol)
	//_, _ = file.Seek(0, 0)

	lastId := uint64(0)
	numPrices := uint32(0)
	date := (*item).Date()

	var err error = nil

	indexNode := fs.index.Lookup(symbol, date)
	for item != nil {
		if indexNode == nil {
			err = errors.New(fmt.Sprintf("Unable to resolve index node for date %d", date))
			break
		}

		if (*item).Date() != date {
			if numPrices != indexNode.GetCount() {
				err = errors.New(fmt.Sprintf("Index count mismatch for date %d", date))
				break
			}
			date = (*item).Date()
			indexNode = fs.index.Lookup(symbol, date)
			numPrices = 1
		} else {
			numPrices++
		}

		if (*item).Id() <= lastId {
			err = errors.New(fmt.Sprintf("Non-increasing ID detected for date %d", date))
			break
		}

		lastId = (*item).Id()
		item = fs.adapter.Read(*file, symbol)
	}

	// one more check on count
	if indexNode != nil && numPrices != indexNode.GetCount() {
		err = errors.New(fmt.Sprintf("Index count mismatch for date %d", indexNode.GetDate()))
	}

	_ = file.Close()
	return err
}
