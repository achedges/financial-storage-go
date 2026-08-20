package buffers

import (
	"time"
)

type FeedForwardBuffer[T any] interface {
	Peek() *T
	Next() *T
	NextAt(instant time.Time) *T
	Prev() *T
	Size() int
	Index() int
}
