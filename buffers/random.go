package buffers

type RandomAccessBuffer[T any] struct {
	elements []T
	size     int
	index    int
}

func NewRandomAccessBufferWithCapacity[T any](capacity int) *RandomAccessBuffer[T] {
	return &RandomAccessBuffer[T]{
		elements: make([]T, 0, capacity),
		size:     capacity,
		index:    -1,
	}
}

func NewRandomAccessBufferWithElements[T any](elements []T) *RandomAccessBuffer[T] {
	return &RandomAccessBuffer[T]{
		elements: elements,
		size:     len(elements),
		index:    len(elements) - 1,
	}
}

func (rab *RandomAccessBuffer[T]) Size() int {
	return rab.size
}

func (rab *RandomAccessBuffer[T]) Index() int {
	return rab.index
}

func (rab *RandomAccessBuffer[T]) Position(offset int) int {
	if offset >= rab.size || len(rab.elements) == 0 {
		return -1
	}
	return (rab.index - offset) % rab.size
}

func (rab *RandomAccessBuffer[T]) Last(offset int) *T {
	pos := rab.Position(offset)
	if pos < 0 {
		return nil
	}
	return &rab.elements[pos]
}

func (rab *RandomAccessBuffer[T]) Add(element T) {
	rab.index++
	if len(rab.elements) < rab.size {
		rab.elements = append(rab.elements, element)
	} else {
		rab.elements[rab.Position(0)] = element
	}
}
