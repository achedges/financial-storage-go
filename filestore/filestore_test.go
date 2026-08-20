package filestore_test

import (
	"os"
	"testing"

	"github.com/achedges/financial-storage-go/filestore"
	"github.com/achedges/go-assertions"
)

func getTestItem(date uint64, time uint32, price float64) SimpleFileStoreItem {
	item := SimpleFileStoreItem{
		date:  date,
		time:  time,
		price: price,
	}
	item.id = item.GenerateId()
	return item
}

func cleanup(adapter *SimpleFileStoreItemAdapter[SimpleFileStoreItem], symbol string) {
	_ = os.Remove(adapter.GetDataFilePath(symbol))
	_ = os.Remove(adapter.GetIndexFilePath(symbol))

}

func TestFileStore_WriteNewItems(t *testing.T) {
	adapter := NewSimpleFileStoreItemAdapter("test_resources", "", "")
	fs := filestore.NewFileStore(adapter)

	// write sample data
	items := []SimpleFileStoreItem{
		getTestItem(20260101, 1, 1.23),
		getTestItem(20260101, 3, 7.89), // ensure records are being sorted
		getTestItem(20260101, 2, 4.56),
	}
	fs.Write("TEST_IO", 20260101, items)

	//  read it back
	readItems := fs.Read("TEST_IO", 20260101, 20260101)
	assertions.EqualInts(3, len(readItems), t)
	assertions.EqualFloats(1.23, readItems[0].price, t)
	assertions.EqualFloats(4.56, readItems[1].price, t)
	assertions.EqualFloats(7.89, readItems[2].price, t)

	// write a new date
	items[0] = getTestItem(20260102, 1, 9.87)
	items[1] = getTestItem(20260102, 2, 6.54)
	items[2] = getTestItem(20260102, 3, 3.21)
	fs.Write("TEST_IO", 20260102, items)

	// read it back
	readItems = fs.Read("TEST_IO", 20260101, 20260102)
	assertions.EqualInts(6, len(readItems), t)
	assertions.EqualFloats(1.23, readItems[0].price, t)
	assertions.EqualFloats(4.56, readItems[1].price, t)
	assertions.EqualFloats(7.89, readItems[2].price, t)
	assertions.EqualFloats(9.87, readItems[3].price, t)
	assertions.EqualFloats(6.54, readItems[4].price, t)
	assertions.EqualFloats(3.21, readItems[5].price, t)

	cleanup(adapter, "TEST_IO")
}

func TestFileStore_WriteExistingItems_Expand(t *testing.T) {
	adapter := NewSimpleFileStoreItemAdapter("test_resources", "", "")
	fs := filestore.NewFileStore(adapter)

	// write sample data
	items := []SimpleFileStoreItem{
		getTestItem(20260101, 1, 1.23),
		getTestItem(20260101, 3, 7.89), // ensure records are being sorted
		getTestItem(20260101, 2, 4.56),
	}
	fs.Write("TEST_INC", 20260101, items)

	// add one more entry
	items = append(items, getTestItem(20260101, 4, 9.99))
	fs.Write("TEST_INC", 20260101, items)

	// read it back
	readItems := fs.Read("TEST_INC", 20260101, 20260101)
	assertions.EqualInts(4, len(readItems), t)
	assertions.EqualFloats(1.23, readItems[0].price, t)
	assertions.EqualFloats(4.56, readItems[1].price, t)
	assertions.EqualFloats(7.89, readItems[2].price, t)
	assertions.EqualFloats(9.99, readItems[3].price, t)

	fileInfo, _ := os.Stat(adapter.GetDataFilePath("TEST_INC"))
	assertions.EqualUints(adapter.GetRecordSizeBytes(4), uint32(fileInfo.Size()), t)

	cleanup(adapter, "TEST_INC")
}

func TestFileStore_WriteExistingItems_Compact(t *testing.T) {
	adapter := NewSimpleFileStoreItemAdapter("test_resources", "", "")
	fs := filestore.NewFileStore(adapter)

	// write sample data for date 1
	items := []SimpleFileStoreItem{
		getTestItem(20260101, 1, 1.1),
		getTestItem(20260101, 3, 3.3), // ensure records are being sorted
		getTestItem(20260101, 2, 2.2),
	}
	fs.Write("TEST_INC", 20260101, items)

	// write sample data for date 2
	items2 := []SimpleFileStoreItem{
		getTestItem(20260102, 1, 4.4),
		getTestItem(20260102, 2, 5.5),
		getTestItem(20260102, 3, 6.6),
	}
	fs.Write("TEST_INC", 20260102, items2)

	// remove an entry
	updatedItems := []SimpleFileStoreItem{
		getTestItem(20260101, 1, 1.1),
		getTestItem(20260101, 2, 2.2),
	}
	fs.Write("TEST_INC", 20260101, updatedItems)

	// read it back
	readItems := fs.Read("TEST_INC", 20260101, 20260102)
	assertions.EqualInts(5, len(readItems), t)
	assertions.EqualFloats(1.1, readItems[0].price, t)
	assertions.EqualFloats(2.2, readItems[1].price, t)
	assertions.EqualFloats(4.4, readItems[2].price, t)
	assertions.EqualFloats(5.5, readItems[3].price, t)
	assertions.EqualFloats(6.6, readItems[4].price, t)

	fileInfo, _ := os.Stat(adapter.GetDataFilePath("TEST_INC"))
	assertions.EqualUints(adapter.GetRecordSizeBytes(5), uint32(fileInfo.Size()), t)

	cleanup(adapter, "TEST_INC")
}

func TestFileStore_CheckIntegrity_NoErrors(t *testing.T) {
	adapter := NewSimpleFileStoreItemAdapter("test_resources", "", "")
	fs := filestore.NewFileStore(adapter)

	// write records for a date
	items := []SimpleFileStoreItem{
		getTestItem(20260101, 1, 1.1),
		getTestItem(20260101, 2, 2.2),
		getTestItem(20260101, 3, 3.3),
	}
	fs.Write("TEST_NO_ERR", 20260101, items)
	err := fs.CheckIntegrity("TEST_NO_ERR")
	assertions.True(err == nil, t)

	cleanup(adapter, "TEST_NO_ERR")
}

func TestFileStore_CheckIntegrity_UnresolvedIndexNode(t *testing.T) {
	adapter := NewSimpleFileStoreItemAdapter("test_resources", "", "")
	fs := filestore.NewFileStore(adapter)

	// write records for a date
	items := []SimpleFileStoreItem{
		getTestItem(20260101, 1, 1.1),
		getTestItem(20260101, 2, 2.2),
	}
	fs.Write("TEST_UNRES_NODE", 20260101, items)

	// reset the index file
	_ = os.Truncate(adapter.GetIndexFilePath("TEST_UNRES_NODE"), 0)

	// check integrity using a new FileStore instance
	checkFs := filestore.NewFileStore(adapter)
	err := checkFs.CheckIntegrity("TEST_UNRES_NODE")
	if err == nil {
		t.Error("Expected unresolved date index node error")
		t.FailNow()
	}

	assertions.EqualStrings("Unable to resolve index node for date 20260101", err.Error(), t)

	cleanup(adapter, "TEST_UNRES_NODE")
}

func TestFileStore_CheckIntegrity_IndexCountMismatch(t *testing.T) {
	adapter := NewSimpleFileStoreItemAdapter("test_resources", "", "")
	fs := filestore.NewFileStore(adapter)

	// write records for a date
	items := []SimpleFileStoreItem{
		getTestItem(20260101, 1, 1.1),
		getTestItem(20260101, 2, 2.2),
	}
	fs.Write("TEST_COUNT", 20260101, items)

	// modify the index file
	_ = os.WriteFile(adapter.GetIndexFilePath("TEST_COUNT"), []byte("20260101,0,1"), 644)

	// check integrity using a new FileStore instance
	checkFs := filestore.NewFileStore(adapter)
	err := checkFs.CheckIntegrity("TEST_COUNT")
	if err == nil {
		t.Error("Expected index count mismatch error")
		t.FailNow()
	}

	assertions.EqualStrings("Index count mismatch for date 20260101", err.Error(), t)

	cleanup(adapter, "TEST_COUNT")
}

func TestFileStore_CheckIntegrity_NonIncreasingId(t *testing.T) {
	adapter := NewSimpleFileStoreItemAdapter("test_resources", "", "")
	fs := filestore.NewFileStore(adapter)

	// write records for a date
	items := []SimpleFileStoreItem{
		getTestItem(20260101, 1, 1.1),
		getTestItem(20260101, 2, 2.2),
		getTestItem(20260101, 2, 2.2),
	}
	fs.Write("TEST_COUNT", 20260101, items)

	// check integrity using a new FileStore instance
	err := fs.CheckIntegrity("TEST_COUNT")
	if err == nil {
		t.Error("Expected non-increasing ID error")
		t.FailNow()
	}

	assertions.EqualStrings("Non-increasing ID detected for date 20260101", err.Error(), t)

	cleanup(adapter, "TEST_COUNT")
}
