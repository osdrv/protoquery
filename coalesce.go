package protoquery

import (
	reflect "reflect"

	"google.golang.org/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
)

func toMessage(v protoreflect.Value) (protoreflect.Message, bool) {
	if msg, ok := v.Interface().(protoreflect.Message); ok {
		return msg, ok
	}
	return nil, false
}

func isMessage(v protoreflect.Value) bool {
	_, ok := toMessage(v)
	return ok
}

func toList(v protoreflect.Value) (protoreflect.List, bool) {
	if list, ok := v.Interface().(protoreflect.List); ok {
		return list, ok
	}
	return nil, false
}

func isList(v protoreflect.Value) bool {
	_, ok := toList(v)
	return ok
}

func toMap(v protoreflect.Value) (protoreflect.Map, bool) {
	if list, ok := v.Interface().(protoreflect.Map); ok {
		return list, ok
	}
	return nil, false
}

func isMap(v protoreflect.Value) bool {
	_, ok := toMap(v)
	return ok
}

func toBytes(v protoreflect.Value) ([]byte, bool) {
	if bytes, ok := v.Interface().([]byte); ok {
		return bytes, ok
	}
	return nil, false
}

func isBytes(v protoreflect.Value) bool {
	_, ok := toBytes(v)
	return ok
}

func findFieldByName(msg proto.Message, name string) (protoreflect.FieldDescriptor, bool) {
	fields := msg.ProtoReflect().Descriptor().Fields()
	fd := fields.ByName(protoreflect.Name(name))
	return fd, fd != nil
}

func toBool(v any) (bool, error) {
	return reflect.ValueOf(v).Bool(), nil
}

func toInt64(v any) (int64, error) {
	return reflect.ValueOf(v).Int(), nil
}

func toFloat64(v any) (float64, error) {
	return reflect.ValueOf(v).Float(), nil
}

func castToProtoreflectKind(v any, kind protoreflect.Kind) (any, bool) {
	switch v.(type) {
	case bool:
		switch kind {
		case protoreflect.BoolKind:
			return v, true
		default:
			return nil, false
		}
	case string:
		switch kind {
		case protoreflect.StringKind:
			return v, true
		case protoreflect.BytesKind:
			return []byte(v.(string)), true
		default:
			return nil, false
		}
	case []byte:
		switch kind {
		case protoreflect.StringKind:
			return string(v.([]byte)), true
		case protoreflect.BytesKind:
			return v, true
		default:
			return nil, false
		}
	case int:
		switch kind {
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
			return int32(v.(int)), true
		case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
			return int64(v.(int)), true
		case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
			return uint32(v.(int)), true
		case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
			return uint64(v.(int)), true
		default:
			return nil, false
		}
	case int32:
		switch kind {
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
			return int32(v.(int32)), true
		case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
			return int64(v.(int32)), true
		case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
			return uint32(v.(int32)), true
		case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
			return uint64(v.(int32)), true
		default:
			return nil, false
		}
	case int64:
		switch kind {
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
			return int32(v.(int64)), true
		case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
			return int64(v.(int64)), true
		case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
			return uint32(v.(int64)), true
		case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
			return uint64(v.(int64)), true
		default:
			return nil, false
		}
	case uint:
		switch kind {
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
			return int32(v.(uint)), true
		case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
			return int64(v.(uint)), true
		case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
			return uint32(v.(uint)), true
		case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
			return uint64(v.(uint)), true
		default:
			return nil, false
		}
	case uint32:
		switch kind {
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
			return int32(v.(uint32)), true
		case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
			return int64(v.(uint32)), true
		case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
			return uint32(v.(uint32)), true
		case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
			return uint64(v.(uint32)), true
		default:
			return nil, false
		}
	case uint64:
		switch kind {
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
			return int32(v.(uint64)), true
		case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
			return int64(v.(uint64)), true
		case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
			return uint32(v.(uint64)), true
		case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
			return uint64(v.(uint64)), true
		default:
			return nil, false
		}
	// TODO(osdrv): implement me
	default:
		return nil, false
	}
}
