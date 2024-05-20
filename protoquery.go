package protoquery

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type ProtoQuery struct {
	query Query
}

func Compile(q string) (*ProtoQuery, error) {
	tokens, err := tokenizeXPathQuery(q)
	if err != nil {
		return nil, err
	}
	query, err := compileQuery(tokens)
	if err != nil {
		return nil, err
	}
	return &ProtoQuery{
		query: query,
	}, nil
}

func (pq *ProtoQuery) Find(root proto.Message) (proto.Message, error) {
	panic("not implemented")
}

func (pq *ProtoQuery) FindAll(msg proto.Message) []proto.Message {
	res := []proto.Message{}

	type queueItem struct {
		qix int
		ptr proto.Message
	}

	queue := []queueItem{{0, msg}}
	var head queueItem
	for len(queue) > 0 {
		head, queue = queue[0], queue[1:]
		if head.qix >= len(pq.query)-1 {
			res = append(res, head.ptr)
			continue
		}
		qstep := pq.query[head.qix]
		switch qstep.Kind() {
		case RootQueryStepKind:
			queue = append(queue, queueItem{
				head.qix + 1,
				head.ptr,
			})
		case NodeQueryStepKind:
			field := head.ptr.ProtoReflect().Descriptor().Fields().ByName(
				protoreflect.Name(qstep.Name()),
			)
			if field == nil {
				// Field not found
				continue
			}
			if field.IsList() {
				list := head.ptr.ProtoReflect().Get(field).List()
				// If the next query step is an index, no need to populate all items
				// in the queue, just the one at the index.
				for i := 0; i < list.Len(); i++ {
					if pq.query[head.qix+1].Kind() == NodeQueryStepKind {
						if !pq.query[head.qix+1].Predicate().IsMatch(i, list.Get(i).Message().Interface()) {
							continue
						}
					}
					queue = append(queue, queueItem{
						head.qix + 1,
						list.Get(i).Message().Interface(),
					})
				}
			} else {
				queue = append(queue, queueItem{
					head.qix + 1,
					head.ptr.ProtoReflect().Get(field).Message().Interface(),
				})
			}
		default:
			panic("not implemented")
		}
	}

	return res
}
