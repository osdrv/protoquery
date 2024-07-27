package protoquery

type Serializable[K comparable] interface {
	Serialize() (K, bool)
}

type Queue[K comparable, T Serializable[K]] struct {
	items []T
	memo  map[K]struct{}
}

func NewQueue[K comparable, T Serializable[K]]() *Queue[K, T] {
	return &Queue[K, T]{
		items: make([]T, 0, 1),
		memo:  make(map[K]struct{}),
	}
}

func (q *Queue[K, T]) Len() int {
	return len(q.items)
}

func (q *Queue[K, T]) Push(item T) {
	q.items = append(q.items, item)
}

func (q *Queue[K, T]) PushUniq(item T) {
	if k, ok := item.Serialize(); ok {
		if _, ok := q.memo[k]; ok {
			return
		}
		q.memo[k] = struct{}{}
	}
	// if we can't serialize, we can't check for uniqueness so we just add it.
	q.items = append(q.items, item)
}

func (q *Queue[K, T]) Pop() T {
	item := q.items[0]
	q.items = q.items[1:]
	return item
}
