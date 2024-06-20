package protoquery

import (
	"fmt"

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

// queueItem is an internal structure to keep track of the moving multi-head pointer.
type queueItem struct {
	qix int
	ptr protoreflect.Value
}

func debugf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func panicf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

func isList(val protoreflect.Value) bool {
	if !val.IsValid() {
		return false
	}
	_, ok := val.Interface().(protoreflect.List)
	return ok
}

func isBytes(val protoreflect.Value) bool {
	if !val.IsValid() {
		return false
	}
	_, ok := val.Interface().([]byte)
	return ok
}

func flatten(val protoreflect.Value) []interface{} {
	res := []interface{}{}
	switch val.Interface().(type) {
	case protoreflect.Message:
		res = append(res, val.Message().Interface())
	case protoreflect.List:
		list := val.List()
		for i := 0; i < list.Len(); i++ {
			res = append(res, flatten(list.Get(i))...)
		}
	default:
		// giving up, returning the value as is
		res = append(res, val.Interface())
	}
	return res
}

func (pq *ProtoQuery) FindAll(root proto.Message) []interface{} {
	res := []interface{}{}
	if root == nil {
		return res
	}
	queue := []queueItem{{
		qix: 0,
		ptr: protoreflect.ValueOf(root.ProtoReflect()),
	}}
	var head queueItem
	for len(queue) > 0 {
		head, queue = queue[0], queue[1:]
		if head.qix >= len(pq.query) {
			res = append(res, flatten(head.ptr)...)
			continue
		}
		step := pq.query[head.qix]
		switch step.Kind() {
		case RootQueryStepKind:
			debugf("Root step: %s", step)
			queue = append(queue, queueItem{
				qix: head.qix + 1,
				ptr: head.ptr,
			})
		case NodeQueryStepKind:
			debugf("Node step: %s", step)
			cs := []protoreflect.Value{}
			if isList(head.ptr) {
				for i := 0; i < head.ptr.List().Len(); i++ {
					cs = append(cs, head.ptr.List().Get(i))
				}
			} else {
				cs = append(cs, head.ptr)
			}
			for _, c := range cs {
				msg := c.Message()
				field, ok := findFieldByName(msg.Interface(), step.(*NodeQueryStep).name)
				if !ok {
					continue
				}
				queue = append(queue, queueItem{
					qix: head.qix + 1,
					ptr: msg.Get(field),
				})
			}
		case KeyQueryStepKind:
			debugf("Key step: %s", step)
			panic("TODO(osdrv): not implemented")
		//case AttrFilterQueryStepKind:
		//	debugf("Attr step: %s", step)
		//	match := []protoreflect.Value{}
		//	afs := step.(*AttrFilterQueryStep)
		//	if isList(head.ptr) {
		//		for i := 0; i < head.ptr.List().Len(); i++ {
		//			val := head.ptr.List().Get(i)
		//			if afs.Match(val) {
		//				match = append(match, val)
		//			}
		//		}
		//	} else {
		//		if afs.Match(head.ptr) {
		//			match = append(match, head.ptr)
		//		}
		//	}
		//	for _, val := range match {
		//		queue = append(queue, queueItem{
		//			qix: head.qix + 1,
		//			ptr: val,
		//		})
		//	}
		// TODO
		default:
			panicf("Query step kind %+v is not supported", step.Kind())
		}
	}
	return res
}

func findFieldByName(msg proto.Message, name string) (protoreflect.FieldDescriptor, bool) {
	fields := msg.ProtoReflect().Descriptor().Fields()
	fd := fields.ByName(protoreflect.Name(name))
	return fd, fd != nil
}
