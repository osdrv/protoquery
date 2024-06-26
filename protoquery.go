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
	qix   int
	ptr   protoreflect.Value
	descr protoreflect.FieldDescriptor
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
				qix:   head.qix + 1,
				ptr:   head.ptr,
				descr: head.descr,
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
				if !c.IsValid() {
					continue
				}
				msg := c.Message()
				fd, ok := findFieldByName(msg.Interface(), step.(*NodeQueryStep).name)
				if !ok {
					continue
				}
				queue = append(queue, queueItem{
					qix:   head.qix + 1,
					ptr:   msg.Get(fd),
					descr: fd,
				})
			}
		case KeyQueryStepKind:
			ks := step.(*KeyQueryStep)
			if isList(head.ptr) {
				list := head.ptr.List()
				ctx := NewEvalContext(list)
				// TODO(osdrv): we can pre-compute this as a property of the query
				// rather than re-computing it on the go.
				// isAllPropertyExprs would check if the key only consists of
				// attribute properties. I.e. it only checks if these properties
				// are present in the message.
				// E.g. [@foo && @bar && @baz]
				// TODO(osdrv): all props + bool checks is still boolean.
				// E.g. [@foo && @bar='value' && true]
				enforceBool := isAllPropertyExprs(ks.expr)
				ctx = ctx.Copy(WithEnforceBool(enforceBool))
				typ, err := ks.expr.Type(ctx)
				if err != nil {
					debugf("keyStep.Type(list) returned an error: %s", err)
					continue
				}
				switch typ {
				case TypeBool:
					// 1. Initialize a new list to store the intermediate results.
					// 2. The list should have the same signature as the original list.
					// 3. Populate the new list with the matching elements.
					// 4. Append the new list to the queue.
					tl := NewTmpList(head.descr)
					for i := 0; i < list.Len(); i++ {
						ctxel := NewIndexedEvalContext(
							list.Get(i).Interface(),
							i,
							WithEnforceBool(enforceBool),
						)
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
							tl.Append(list.Get(i))
						}
					}
					if tl.Len() > 0 {
						queue = append(queue, queueItem{
							qix:   head.qix + 1,
							ptr:   protoreflect.ValueOf(tl),
							descr: head.descr, // The type should not change: we are still in a list.
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
							// TODO(osdrv): type descriptor
						})
					}
				default:
					debugf("keyStep.Type(list) returned an unsupported type: %s", typ)
					continue
				}
			} else if isMap(head.ptr) {
				mp := head.ptr.Map()
				ctx := NewEvalContext(mp)
				k, err := ks.expr.Eval(ctx)
				if err != nil {
					debugf("keyStep.Eval(map) returned an error: %s", err)
					continue
				}
				if head.descr == nil {
					debugf("No information about map key type, trying the raw value")
				} else {
					if fd := head.descr; fd == nil || !fd.IsMap() {
						panicf("Unexpected descriptor kind: want protoreflect.Map, got %v", head.descr.Kind())
					}
					var ok bool
					mkKind := head.descr.MapKey().Kind()
					k, ok = castToProtoreflectKind(k, mkKind)
					if !ok {
						debugf("Can not cast value %+v to protoreflect.Kind=%v", k, mkKind)
						continue
					}
				}
				exprval := protoreflect.ValueOf(k)
				key := exprval.MapKey()
				if mp.Has(key) {
					queue = append(queue, queueItem{
						qix:   head.qix + 1,
						ptr:   mp.Get(key),
						descr: head.descr.MapValue(),
					})
				}

				//} else if isMessage(head.ptr) {

			} else if isBytes(head.ptr) {
				ctx := NewEvalContext(head.ptr)
				typ, err := ks.expr.Type(ctx)
				if err != nil {
					debugf("keyStep.Type(bytes) returned an error: %s", err)
					continue
				}
				if typ != TypeInt {
					debugf("keyStep.Type(bytes) returned an unsupported type: %s", typ)
					continue
				}
				k, err := ks.expr.Eval(ctx)
				if err != nil {
					debugf("keyStep.Eval(bytes) returned an error: %s", err)
					continue
				}
				ix, err := toInt64(k)
				if err != nil {
					debugf("keyStep.Eval(bytes) returned an error on toInt64: %s", err)
					continue
				}
				bytes := head.ptr.Bytes()
				if ix >= 0 && int(ix) < len(bytes) {
					queue = append(queue, queueItem{
						qix: head.qix + 1,
						// protoreflect does not support any ints below 32bits, hence the type casting
						ptr: protoreflect.ValueOf(uint32(bytes[ix])),
						// TODO(osdrv): type descriptor
					})
				}
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
