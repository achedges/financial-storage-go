package buffers

import (
	"testing"

	"github.com/achedges/go-assertions"
)

func TestRandomAccessBuffer_Init_Elements(t *testing.T) {
	buffer := NewRandomAccessBufferWithElements([]int{0, 1, 2, 3, 4})
	assertions.EqualInts(5, buffer.Size(), t)
	assertions.EqualInts(4, buffer.Index(), t)
}

func TestRandomAccessBuffer_Init_capacity(t *testing.T) {
	buffer := NewRandomAccessBufferWithCapacity[int](5)
	assertions.EqualInts(5, buffer.Size(), t)
	assertions.EqualInts(-1, buffer.Index(), t)
}

func TestRandomAccessBuffer_Position(t *testing.T) {
	buffer := NewRandomAccessBufferWithElements([]int{0, 1, 2})
	assertions.EqualInts(2, buffer.Position(0), t)
	assertions.EqualInts(1, buffer.Position(1), t)
	assertions.EqualInts(0, buffer.Position(2), t)
	assertions.EqualInts(-1, buffer.Position(3), t)
	assertions.EqualInts(-1, buffer.Position(4), t)
}

func TestRandomAccessBuffer_Last(t *testing.T) {
	buffer := NewRandomAccessBufferWithElements([]int{0, 1, 2})
	assertions.EqualInts(2, *buffer.Last(0), t)
	assertions.EqualInts(1, *buffer.Last(1), t)
	assertions.EqualInts(0, *buffer.Last(2), t)
	assertions.True(buffer.Last(4) == nil, t)
}

func TestRandomAccessBuffer_Add(t *testing.T) {
	buffer := NewRandomAccessBufferWithElements([]int{0, 1, 2})

	buffer.Add(3)
	assertions.EqualInts(3, *buffer.Last(0), t)
	assertions.EqualInts(2, *buffer.Last(1), t)
	assertions.EqualInts(1, *buffer.Last(2), t)
	assertions.True(buffer.Last(3) == nil, t)

	buffer.Add(4)
	assertions.EqualInts(4, *buffer.Last(0), t)
	assertions.EqualInts(3, *buffer.Last(1), t)
	assertions.EqualInts(2, *buffer.Last(2), t)
	assertions.True(buffer.Last(3) == nil, t)

	// exercise the wrap-around a bit more
	buffer2 := NewRandomAccessBufferWithElements([]int{0, 1, 2})
	for i := 3; i < 10; i++ {
		buffer2.Add(i)
		assertions.EqualInts(i, *buffer2.Last(0), t)
	}

	buffer3 := NewRandomAccessBufferWithCapacity[int](3)
	buffer3.Add(1)
	buffer3.Add(2)
	buffer3.Add(3)
	assertions.EqualInts(3, buffer3.Size(), t)
	assertions.EqualInts(3, *buffer3.Last(0), t)
}
