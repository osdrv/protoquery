package protoquery

import (
	"fmt"
	"strings"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
)

func debugf(format string, args ...interface{}) {
	if DEBUG {
		fmt.Printf(format+"\n", args...)
	}
}

func panicf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

func printProtoVal(v protoreflect.Value) string {
	if !v.IsValid() {
		return "<invalid>"
	}
	if msg, ok := toMessage(v); ok {
		var b strings.Builder
		b.WriteString(string(msg.Descriptor().Name()))
		b.WriteString("{")
		for i := 0; i < msg.Descriptor().Fields().Len(); i++ {
			fd := msg.Descriptor().Fields().Get(i)
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("%s: %s", fd.Name(), printProtoVal(msg.Get(fd))))
		}
		b.WriteString("}")
		return b.String()
	} else if list, ok := toList(v); ok {
		var b strings.Builder
		b.WriteString("[")
		for i := 0; i < list.Len(); i++ {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(printProtoVal(list.Get(i)))
		}
		b.WriteString("]")
		return b.String()
	} else if mapv, ok := toMap(v); ok {
		var b strings.Builder
		b.WriteString("{")
		mapv.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
			if b.Len() > 1 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("%s: %s", printProtoVal(k.Value()), printProtoVal(v)))
			return true
		})
		b.WriteString("}")
		return b.String()
	} else {
		return fmt.Sprintf("%v", v.Interface())
	}
}
