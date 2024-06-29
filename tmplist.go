package protoquery

import protoreflect "google.golang.org/protobuf/reflect/protoreflect"

type TmpList struct {
	descr    protoreflect.FieldDescriptor
	elements []protoreflect.Value
}

var _ protoreflect.List = (*TmpList)(nil)

func NewTmpList(descr protoreflect.FieldDescriptor) *TmpList {
	return &TmpList{
		descr:    descr,
		elements: make([]protoreflect.Value, 0),
	}
}

func (tl *TmpList) Len() int {
	return len(tl.elements)
}

func (tl *TmpList) Get(i int) protoreflect.Value {
	return tl.elements[i]
}

func (tl *TmpList) Append(v protoreflect.Value) {
	tl.elements = append(tl.elements, v)
}

func (tl *TmpList) Set(i int, v protoreflect.Value) {
	tl.elements[i] = v
}

func (tl *TmpList) Truncate(n int) {
	tl.elements = tl.elements[:n]
}

func (tl *TmpList) AppendMutable() protoreflect.Value {
	v := protoreflect.Value{}
	tl.Append(v)
	return v
}

func (tl *TmpList) NewElement() protoreflect.Value {
	// TODO(osdrv): review this code.
	// This might cause problems because we contain no type data.
	return protoreflect.Value{}
}

func (tl *TmpList) IsValid() bool {
	// Given it is initialized, it is always valid
	return true
}
