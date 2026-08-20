package index

import (
	"fmt"
	"strconv"
	"strings"
)

type Node struct {
	date   uint64
	offset int64
	count  uint32
}

func NewNode(date uint64, offset int64, count uint32) *Node {
	return &Node{
		date:   date,
		offset: offset,
		count:  count,
	}
}

func NewNodeFromCsv(csvLine string) *Node {
	fields := strings.Split(csvLine, ",")
	_date, _dateErr := strconv.Atoi(strings.TrimSpace(fields[0]))
	_offset, _offsetErr := strconv.Atoi(strings.TrimSpace(fields[1]))
	_count, _countErr := strconv.Atoi(strings.TrimSpace(fields[2]))
	if _dateErr != nil || _offsetErr != nil || _countErr != nil {
		fmt.Printf("Unable to parse CSV line to Node: _dateErr=%s / _offsetErr=%s / _countErr=%s", _dateErr, _offsetErr, _countErr)
		return nil
	}
	node := Node{
		date:   uint64(_date),
		offset: int64(_offset),
		count:  uint32(_count),
	}
	return &node
}

func (n *Node) GetDate() uint64 {
	return n.date
}

func (n *Node) GetOffset() int64 {
	return n.offset
}

func (n *Node) GetCount() uint32 {
	return n.count
}

func (n *Node) ToCsv() string {
	return fmt.Sprintf("%d,%d,%d", n.date, n.offset, n.count)
}

func (n *Node) IsAfter(other *Node) bool {
	return n.date > other.date
}

func (n *Node) IsBefore(other *Node) bool {
	return n.date < other.date
}

func (n *Node) SetOffset(offset int64) {
	n.offset = offset
}

func (n *Node) SetCount(count uint32) {
	n.count = count
}

func (n *Node) UpdateOffset(deltabytes int64) {
	if n.offset+deltabytes < 0 {
		n.offset = 0
	} else {
		n.offset += deltabytes
	}
}
