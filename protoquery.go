package protoquery

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

// func (pq *ProtoQuery) Find(root proto.Message) (proto.Message, error) {
// 	panic("not implemented")
// }

// queueItem is an internal structure to keep track of the moving multi-head pointer.
type queueItem struct {
	qix int
	ptr protoreflect.Value
}

func protoListToInterfaceList(pl protoreflect.List) interface{} {
	list := make([]interface{}, 0, pl.Len())
	for i := 0; i < pl.Len(); i++ {
		list = append(list, pl.Get(i).Interface())
	}
	return list
}

func (pq *ProtoQuery) FindAll(msg proto.Message) []interface{} {
	res := []interface{}{}
	if msg == nil {
		return res
	}
	queue := []queueItem{{0, protoreflect.ValueOf(msg.ProtoReflect())}}
	var head queueItem
	for len(queue) > 0 {
		head, queue = queue[0], queue[1:]
		if head.qix > len(pq.query)-1 {
			val := head.ptr.Interface()
			// TODO(osdrv): this explicit branching is error-prone.
			// I need to figure out a better way to discriminate between
			// messages and scalars.
			// TODO(osdrv): most likely, this is a good candidate for a interface-based switch.
			if _, ok := val.(protoreflect.Message); ok {
				res = append(res, head.ptr.Message().Interface())
			} else if _, ok := val.(protoreflect.List); ok {
				res = append(res, protoListToInterfaceList(head.ptr.List()))
			} else {
				res = append(res, val)
			}
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
			if field, ok := findFieldByName(head.ptr.Message().Interface(), qstep.Name()); ok {
				// NOTE(osdrv):
				// This is an attempt of a local optimization.
				// XPath path contains the name of the element and we're trying to jump straight to the next step.
				// It won't make sense when this is the last step, hence we need to check the bounds.
				if field.IsList() && head.qix < len(pq.query)-1 {
					list := head.ptr.Message().Get(field).List()
					// If the next query step is an index, no need to populate all items
					// in the queue, just the one at the index.
					for i := 0; i < list.Len(); i++ {
						if pq.query[head.qix+1].Kind() == NodeQueryStepKind {
							element := list.Get(i).Message().Interface()
							if !nameMatches(element.ProtoReflect().Descriptor().Name(), pq.query[head.qix+1].Name()) {
								continue
							}
							if !pq.query[head.qix+1].Predicate().IsMatch(i, element) {
								continue
							}
						}
						queue = append(queue, queueItem{
							head.qix + 2,
							list.Get(i),
						})
					}
				} else {
					nextfield := head.ptr.Message().Get(field)
					nextval := protoreflect.ValueOf(nextfield.Interface())
					queue = append(queue, queueItem{
						head.qix + 1,
						nextval,
					})
				}
			} else if oneoff, ok := findOneOfByName(head.ptr.Message().Interface(), qstep.Name()); ok {
				fmt.Printf("oneoff: %s", oneoff)
			} else {
				continue
			}
		default:
			panic("not implemented")
		}
	}

	return res
}

func snakeCase(s string) string {
	var b bytes.Buffer
	csr := cases.Title(language.English)
	for _, ss := range strings.Split(s, "_") {
		b.WriteString(csr.String(ss))
	}
	return b.String()
}

func nameMatches(name protoreflect.Name, query string) bool {
	return string(name) == snakeCase(query)
}

func findOneOfByName(msg proto.Message, name string) (protoreflect.OneofDescriptor, bool) {
	od := msg.ProtoReflect().Descriptor().Oneofs().ByName(protoreflect.Name(name))
	return od, od != nil
}

func findFieldByName(msg proto.Message, name string) (protoreflect.FieldDescriptor, bool) {
	fields := msg.ProtoReflect().Descriptor().Fields()
	fd := fields.ByName(protoreflect.Name(name))
	return fd, fd != nil
}
