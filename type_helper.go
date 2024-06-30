package protoquery

import (
	"google.golang.org/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
)

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

func findFieldByName(msg proto.Message, name string) (protoreflect.FieldDescriptor, bool) {
	fields := msg.ProtoReflect().Descriptor().Fields()
	fd := fields.ByName(protoreflect.Name(name))
	return fd, fd != nil
}
