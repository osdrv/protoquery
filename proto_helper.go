package protoquery

import protoreflect "google.golang.org/protobuf/reflect/protoreflect"

func enumStr(fd protoreflect.FieldDescriptor, v protoreflect.Value) (string, bool) {
	ed := fd.Enum()
	vv := ed.Values()
	i := int(v.Enum())
	if i >= 0 && i < vv.Len() {
		return string(vv.Get(i).Name()), true
	}
	return "", false
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
