package protoquery

import protoreflect "google.golang.org/protobuf/reflect/protoreflect"

func castToProtoreflectKind(v any, kind protoreflect.Kind) (any, bool) {
	switch v.(type) {
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
