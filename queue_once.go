package protoquery

type Serializable[K comparable] interface {
	Serialize() (K, bool)
}

type QueueOnce[K comparable, T Serializable[K]] struct {
	items []T
	memo  map[K]struct{}
}

func NewQueueOnce[K comparable, T Serializable[K]]() *QueueOnce[K, T] {
	return &QueueOnce[K, T]{
		items: make([]T, 0, 1),
		memo:  make(map[K]struct{}),
	}
}

func (q *QueueOnce[K, T]) Len() int {
	return len(q.items)
}

// Note: QueueOnce memo is never flushed. This could be a problem if
// a highamount of items being enqueued and the key cardinality is high.
func (q *QueueOnce[K, T]) Push(item T) {
	if k, ok := item.Serialize(); ok {
		if _, ok := q.memo[k]; ok {
			return
		}
		q.memo[k] = struct{}{}
	}
	// if we can't serialize, we can't check for uniqueness so we just add it.
	q.items = append(q.items, item)
}

func (q *QueueOnce[K, T]) Pop() T {
	item := q.items[0]
	q.items = q.items[1:]
	return item
}
