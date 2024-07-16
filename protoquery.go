package protoquery

import (
	reflect "reflect"

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

func nameMatch(n protoreflect.Name, f string) bool {
	return f == "*" || string(n) == f
}

func matchMsgFields(m protoreflect.Message, f string) []protoreflect.FieldDescriptor {
	res := []protoreflect.FieldDescriptor{}
	ff := m.Interface().ProtoReflect().Descriptor().Fields()
	for i := 0; i < ff.Len(); i++ {
		if nameMatch(ff.Get(i).Name(), f) {
			res = append(res, ff.Get(i))
		}
	}
	return res
}

func canRecurse(v protoreflect.Value) bool {
	if v.IsValid() {
		switch v.Interface().(type) {
		case protoreflect.Message, protoreflect.List, protoreflect.Map:
			return true
		}
	}
	return false
}

func enumStr(fd protoreflect.FieldDescriptor, v protoreflect.Value) (string, bool) {
	ed := fd.Enum()
	vv := ed.Values()
	i := int(v.Enum())
	if i >= 0 && i < vv.Len() {
		return string(vv.Get(i).Name()), true
	}
	return "", false
}

// flat return a flat list of value(s). If the value is a message, it returns a list with the only element.
// If the value is a list, it its elements.
// The function performs validity check on the value.
func flat(v protoreflect.Value) []protoreflect.Value {
	res := []protoreflect.Value{}
	if isList(v) {
		for i := 0; i < v.List().Len(); i++ {
			if vv := v.List().Get(i); vv.IsValid() {
				res = append(res, v.List().Get(i))
			}
		}
	} else {
		if v.IsValid() {
			res = append(res, v)
		}
	}

	return res
}

// stripProto returns the underlying Go value of the protoreflect.Value.
func stripProto(v protoreflect.Value) any {
	if !v.IsValid() {
		return nil
	}
	switch v.Interface().(type) {
	case protoreflect.Message:
		return v.Message().Interface()
	default:
		return v.Interface()
	}
}

type qmemkey struct {
	qix  int
	uptr uintptr
}

func (pq *ProtoQuery) FindAll(root proto.Message) []any {
	debugf("Query: %s", pq.query)

	res := []any{}
	if root == nil {
		return res
	}
	queue := []queueItem{}

	memo := make(map[qmemkey]bool)
	appendUnique := func(q []queueItem, qi queueItem) []queueItem {
		uptr := uintptr(reflect.ValueOf(qi.ptr).FieldByName("ptr").UnsafePointer())
		k := qmemkey{
			qix:  qi.qix,
			uptr: uptr,
		}
		if DEBUG {
			debugf("schedule map key: %+v", k)
		}
		if uptr == 0 || !canRecurse(qi.ptr) || !memo[k] {
			step := "<OoB>"
			if qi.qix < len(pq.query) {
				step = pq.query[qi.qix].String()
			}
			debugf("scheduled: %s step %s", printProtoVal(qi.ptr), step)
			memo[k] = true
			q = append(q, qi)
			debugf("admitted %+v", k)
		} else {
			debugf("skipped: %+v", k)
		}
		return q
	}

	queue = appendUnique(queue, queueItem{
		qix: 0,
		ptr: protoreflect.ValueOf(root.ProtoReflect()),
	})

	var head queueItem
	for len(queue) > 0 {
		head, queue = queue[0], queue[1:]
		if head.qix >= len(pq.query) {
			for _, v := range flat(head.ptr) {
				res = append(res, stripProto(v))
			}
			continue
		}
		step := pq.query[head.qix]
		if DEBUG {
			debugf("-> current pointer: %s", printProtoVal(head.ptr))
			debugf("~> step: %s", step)
		}
		switch step.Kind() {
		case RootQueryStepKind:
			debugf("Root step: %s", step)
			queue = appendUnique(queue, queueItem{
				qix:   head.qix + 1,
				ptr:   head.ptr,
				descr: head.descr,
			})
		case NodeQueryStepKind:
			debugf("Node step: %s", step)
			for _, c := range flat(head.ptr) {
				if msg, ok := toMessage(c); ok {
					for _, fd := range matchMsgFields(msg, step.(*NodeQueryStep).name) {
						val := msg.Get(fd)
						if fd.Kind() == protoreflect.EnumKind {
							if e, ok := enumStr(fd, val); ok {
								val = protoreflect.ValueOfString(e)
							}
						}
						queue = appendUnique(queue, queueItem{
							qix:   head.qix + 1,
							ptr:   val,
							descr: fd,
						})
					}
				}
			}
		case KeyQueryStepKind:
			ks := step.(*KeyQueryStep)
			if list, ok := toList(head.ptr); ok {
				// TODO(osdrv): we can pre-compute this as a property of the query
				// rather than re-computing it on the go.
				// isAllPropertyExprs would check if the key only consists of
				// attribute properties. I.e. it only checks if these properties
				// are present in the message.
				// E.g. [@foo && @bar && @baz]
				// TODO(osdrv): all props + bool checks is still boolean.
				// E.g. [@foo && @bar='value' && true]
				enforceBool := isAllPropertyExprs(ks.expr)
				ctx := NewEvalContext(list, WithEnforceBool(enforceBool))
				//ctx = ctx.Copy(WithEnforceBool(enforceBool))
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
						queue = appendUnique(queue, queueItem{
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
						queue = appendUnique(queue, queueItem{
							qix: head.qix + 1,
							ptr: list.Get(int(ix)),
							// TODO(osdrv): type descriptor
						})
					}
				default:
					debugf("keyStep.Type(list) returned an unsupported type: %s", typ)
					continue
				}
			} else if mp, ok := toMap(head.ptr); ok {
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
					queue = appendUnique(queue, queueItem{
						qix:   head.qix + 1,
						ptr:   mp.Get(key),
						descr: head.descr.MapValue(),
					})
				}
			} else if bytes, ok := toBytes(head.ptr); ok {
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
				if ix >= 0 && int(ix) < len(bytes) {
					queue = appendUnique(queue, queueItem{
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
		case RecursiveDescentQueryStepKind:
			//debugf("Recursive descent step: %s", step)
			if msg, ok := toMessage(head.ptr); ok {
				// test the message itself
				queue = appendUnique(queue, queueItem{
					qix:   head.qix + 1,
					ptr:   head.ptr,
					descr: head.descr,
				})
				// recurse over all the fields
				for _, fd := range matchMsgFields(msg, "*") {
					if canRecurse(msg.Get(fd)) {
						// preserve the recursive descent query step
						queue = appendUnique(queue, queueItem{
							qix:   head.qix,
							ptr:   msg.Get(fd),
							descr: fd,
						})
					}
					// progress 1 query step forward (drop recursion)
					// queue = appendUnique(queue, queueItem{
					// 	qix:   head.qix + 1,
					// 	ptr:   msg.Get(fd),
					// 	descr: fd,
					// })
				}
			} else if list, ok := toList(head.ptr); ok {
				// progress 1 query step forward (drop recursion)
				//queue = appendUnique(queue, queueItem{
				//	qix:   head.qix + 1,
				//	ptr:   head.ptr,
				//	descr: head.descr,
				//})
				for i := 0; i < list.Len(); i++ {
					if canRecurse(list.Get(i)) {
						// preserve the recursive descent query step
						queue = appendUnique(queue, queueItem{
							qix:   head.qix,
							ptr:   list.Get(i),
							descr: head.descr,
						})
					}
				}
			} else if mp, ok := toMap(head.ptr); ok {
				// progress 1 query step forward (drop recursion)
				//queue = appendUnique(queue, queueItem{
				//	qix:   head.qix + 1,
				//	ptr:   head.ptr,
				//	descr: head.descr,
				//})
				mp.Range(func(key protoreflect.MapKey, value protoreflect.Value) bool {
					if canRecurse(value) {
						// preserve the recursive descent query step
						queue = appendUnique(queue, queueItem{
							qix:   head.qix,
							ptr:   value,
							descr: head.descr.MapValue(),
						})
					}
					return true
				})
			} else {
				debugf("RecursiveDescentQuery is not implemented for %+v", head.ptr.Interface())
			}
		default:
			panicf("Query step %q(kind=%v) is not supported", step.String(), step.Kind())
		}
	}
	return res
}
