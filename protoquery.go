package protoquery

import (
	"os"
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type ProtoQuery struct {
	query Query
}

type qmemkey struct {
	qix  int
	uptr uintptr
}

// queueItem is an internal structure to keep track of the moving multi-head pointer.
type queueItem struct {
	qix   int
	ptr   protoreflect.Value
	descr protoreflect.FieldDescriptor
}

func (qi queueItem) Serialize() (qmemkey, bool) {
	uptr := uintptr(reflect.ValueOf(qi.ptr).FieldByName("ptr").UnsafePointer())
	if uptr == 0 || !canRecurse(qi.ptr) {
		return qmemkey{}, false
	}
	return qmemkey{
		qix:  qi.qix,
		uptr: uptr,
	}, true
}

var (
	DEBUG = os.Getenv("DEBUG") != ""
)

func Compile(q string) (*ProtoQuery, error) {
	tokens, err := tokenizeXPathQuery(q)
	if err != nil {
		return nil, err
	}
	query, err := compileQuery(tokens)
	if err != nil {
		return nil, err
	}
	return &ProtoQuery{query: query}, nil
}

func (pq *ProtoQuery) FindAll(root proto.Message) []any {
	if DEBUG {
		debugf("Query: %s", pq.query)
	}

	res := []any{}
	if root == nil {
		return res
	}

	queue := NewQueue[qmemkey, queueItem]()
	queue.PushUniq(queueItem{
		qix: 0,
		ptr: protoreflect.ValueOf(root.ProtoReflect()),
	})

	var head queueItem
	for queue.Len() > 0 {
		head = queue.Pop()
		// We've reached the end of the query, so we can append the current pointer to the result.
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
			queue.PushUniq(queueItem{
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
						queue.PushUniq(queueItem{
							qix:   head.qix + 1,
							ptr:   val,
							descr: fd,
						})
					}
				} else {
					debugf("Node step: %s: not a message, skipping", step)
				}
			}
		case KeyQueryStepKind:
			debugf("KeyQuery step: %s", step)
			ks := step.(*KeyQueryStep)
			if list, ok := toList(head.ptr); ok {
				// TODO(osdrv): we can pre-compute isAllPropertyExprs as a property
				// of the query, rather than doing it on the go.
				// isAllPropertyExprs would check if the key only consists of
				// attribute properties. I.e. it only checks if these properties
				// are present in the message.
				// E.g. [@foo && @bar && @baz]
				enforceBool := isAllPropertyExprs(ks.expr)
				ctx := NewEvalContext(list, WithEnforceBool(enforceBool))
				typ, err := ks.expr.Type(ctx)
				if err != nil {
					debugf("keyStep.Type(list) returned an error: %s", err)
					continue
				}
				switch typ {
				// Grep mode
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
						if pick, err := toBool(v); err != nil {
							debugf("keyStep.Eval(list):bool returned an error on toBool: %s", err)
							continue
						} else if pick {
							tl.Append(list.Get(i))
						}
					}
					if tl.Len() > 0 {
						queue.PushUniq(queueItem{
							qix:   head.qix + 1,
							ptr:   protoreflect.ValueOf(tl),
							descr: head.descr, // The type descriptor won't change: lists have identical signatures.
						})
					}
				// Index mode
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
					if ix >= 0 && ix < int64(list.Len()) {
						queue.PushUniq(queueItem{
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
						debugf("Unexpected descriptor kind: want protoreflect.Map, got %v", head.descr.Kind())
						continue
					}
					var ok bool
					keyKind := head.descr.MapKey().Kind()
					k, ok = castToProtoreflectKind(k, keyKind)
					if !ok {
						debugf("Can not cast value %+v to protoreflect.Kind=%v", k, keyKind)
						continue
					}
				}
				exprval := protoreflect.ValueOf(k)
				key := exprval.MapKey()
				if mp.Has(key) {
					queue.PushUniq(queueItem{
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
					queue.PushUniq(queueItem{
						qix: head.qix + 1,
						// protoreflect does not support any ints below 32bits, hence the type casting
						ptr: protoreflect.ValueOf(uint32(bytes[ix])),
						// TODO(osdrv): type descriptor
					})
				}
			} else if msg, ok := toMessage(head.ptr); ok {
				// We always enforce bool context on a message.
				ctx := NewEvalContext(msg)
				v, err := ks.expr.Eval(ctx)
				if err != nil {
					debugf("keyStep.Eval(message) returned an error: %s", err)
					continue
				}
				pick, err := toBool(v)
				if err != nil {
					debugf("keyStep.Eval(message) returned an error on toBool: %s", err)
					continue
				}
				if pick {
					queue.PushUniq(queueItem{
						qix:   head.qix + 1,
						ptr:   head.ptr,
						descr: head.descr,
					})
				}
			} else {
				debugf("Current pointer: %s", printProtoVal(head.ptr))
				debugf("Key step: %s", step)
				panic("TODO(osdrv): not implemented")
			}
		case RecursiveDescentQueryStepKind:
			debugf("Recursive descent step: %s", step)
			if msg, ok := toMessage(head.ptr); ok {
				// test the message itself
				queue.PushUniq(queueItem{
					qix:   head.qix + 1,
					ptr:   head.ptr,
					descr: head.descr,
				})
				// recurse over all the fields
				for _, fd := range matchMsgFields(msg, "*") {
					if canRecurse(msg.Get(fd)) {
						// preserve the recursive descent query step
						queue.PushUniq(queueItem{
							qix:   head.qix,
							ptr:   msg.Get(fd),
							descr: fd,
						})
					}
				}
			} else if list, ok := toList(head.ptr); ok {
				for i := 0; i < list.Len(); i++ {
					if canRecurse(list.Get(i)) {
						// preserve the recursive descent query step
						queue.PushUniq(queueItem{
							qix:   head.qix,
							ptr:   list.Get(i),
							descr: head.descr,
						})
					}
				}
			} else if mp, ok := toMap(head.ptr); ok {
				mp.Range(func(key protoreflect.MapKey, value protoreflect.Value) bool {
					if canRecurse(value) {
						// preserve the recursive descent query step
						queue.PushUniq(queueItem{
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
