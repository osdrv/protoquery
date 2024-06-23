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
	query, err := CompileQuery(tokens)
	if err != nil {
		return nil, err
	}
	return &ProtoQuery{
		query: query,
	}, nil
}

// queueItem is an internal structure to keep track of the moving multi-head pointer.
type queueItem struct {
	qix     int
	ptr     protoreflect.Value
	tmplist []protoreflect.Value
}

func isMessage(val protoreflect.Value) bool {
	if !val.IsValid() {
		return false
	}
	_, ok := val.Interface().(protoreflect.Message)
	return ok
}

func isList(val protoreflect.Value) bool {
	if !val.IsValid() {
		return false
	}
	_, ok := val.Interface().(protoreflect.List)
	return ok
}

func isMap(val protoreflect.Value) bool {
	if !val.IsValid() {
		return false
	}
	_, ok := val.Interface().(protoreflect.Map)
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
			if head.tmplist != nil {
				for _, v := range head.tmplist {
					res = append(res, flatten(v)...)
				}
			} else {
				res = append(res, flatten(head.ptr)...)
			}
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
			ks := step.(*KeyQueryStep)
			if isList(head.ptr) {
				list := head.ptr.List()
				ctx := NewEvalContext(list)
				typ, err := ks.expr.Type(ctx)
				if err != nil {
					debugf("keyStep.Type(list) returned an error: %s", err)
					continue
				}
				switch typ {
				case TypeBool:
					var tl []protoreflect.Value
					for i := 0; i < list.Len(); i++ {
						ctxel := NewEvalContext(list.Get(i).Interface())
						v, err := ks.expr.Eval(ctxel)
						if err != nil {
							debugf("keyStep.Eval(list):bool returned an error on Eval: %s", err)
							continue
						}
						pick, err := toBool(v)
						if err != nil {
							debugf("keyStep.Eval(list):bool returned an error on toBool: %s", err)
							continue
						}
						if pick {
							tl = append(tl, list.Get(i))
						}
					}
					if len(tl) > 0 {
						queue = append(queue, queueItem{
							qix:     head.qix + 1,
							tmplist: tl,
						})
					}
				case TypeInt:
					v, err := ks.expr.Eval(ctx)
					if err != nil {
						debugf("keyStep.Eval(list):int returned an error on Eval: %s", err)
						continue
					}
					ix, err := toInt64(v)
					if err != nil {
						debugf("keyStep.Eval(list):int returned an error on toInt64: %s", err)
						continue
					}
					if ix >= 0 && ix <= int64(list.Len()) {
						queue = append(queue, queueItem{
							qix: head.qix + 1,
							ptr: list.Get(int(ix)),
						})
					}
				default:
					debugf("keyStep.Type(list) returned an unsupported type: %s", typ)
					continue
				}
				//} else if isMap(head.ptr) {

				//} else if isMessage(head.ptr) {

				//} else if isBytes(head.ptr) {
			} else {
				debugf("Key step: %s", step)
				panic("TODO(osdrv): not implemented")
			}
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
