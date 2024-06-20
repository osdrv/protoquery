package protoquery

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type QueryStepKind int

const (
	SelfQueryStepKind QueryStepKind = iota
	NodeQueryStepKind
	AttrFilterQueryStepKind
	IndexQueryStepKind
	KeyQueryStepKind
	RootQueryStepKind
	RecursiveDescentQueryStepKind
)

type Query []QueryStep

func (q Query) String() string {
	var s strings.Builder
	for _, step := range q {
		s.WriteString(step.String())
	}
	return s.String()
}

type QueryStep interface {
	Kind() QueryStepKind
	String() string
}

type defaultQueryStep struct{}

func (qs *defaultQueryStep) Predicate() Predicate {
	return nil
}

type SelfQueryStep struct {
	*defaultQueryStep
}

var _ QueryStep = (*SelfQueryStep)(nil)

func (qs *SelfQueryStep) String() string {
	return "."
}

func (qs *SelfQueryStep) Kind() QueryStepKind {
	return SelfQueryStepKind
}

type NodeQueryStep struct {
	*defaultQueryStep
	name string
}

var _ QueryStep = (*NodeQueryStep)(nil)

func (qs *NodeQueryStep) String() string {
	return qs.name
}

func (qs *NodeQueryStep) Kind() QueryStepKind {
	return NodeQueryStepKind
}

type RootQueryStep struct {
	*defaultQueryStep
}

var _ QueryStep = (*RootQueryStep)(nil)

func (qs *RootQueryStep) String() string {
	return "/"
}

func (qs *RootQueryStep) Kind() QueryStepKind {
	return RootQueryStepKind
}

type RecursiveDescentQueryStep struct {
	*defaultQueryStep
}

var _ QueryStep = (*RecursiveDescentQueryStep)(nil)

func (qs *RecursiveDescentQueryStep) String() string {
	return "//"
}

func (qs *RecursiveDescentQueryStep) Kind() QueryStepKind {
	return RecursiveDescentQueryStepKind
}

type AttrFilterQueryStep struct {
	*defaultQueryStep
	predicate Predicate
}

var _ QueryStep = (*AttrFilterQueryStep)(nil)

func (qs *AttrFilterQueryStep) Match(val protoreflect.Value) bool {
	return qs.predicate.Match(val)
}

func (qs *AttrFilterQueryStep) String() string {
	return "[@" + qs.predicate.String() + "]"
}

func (qs *AttrFilterQueryStep) Kind() QueryStepKind {
	return AttrFilterQueryStepKind
}

// Deprecated: use KeyQueryStep instead.
type IndexQueryStep struct {
	*defaultQueryStep
	index int
}

var _ QueryStep = (*IndexQueryStep)(nil)

func (qs *IndexQueryStep) GetElement(list protoreflect.List) (protoreflect.Value, bool) {
	if list.IsValid() && qs.index < list.Len() {
		return list.Get(qs.index), true
	}
	return protoreflect.Value{}, false
}

func (qs *IndexQueryStep) String() string {
	return fmt.Sprintf("[%d]", qs.index)
}

func (qs *IndexQueryStep) Kind() QueryStepKind {
	return IndexQueryStepKind
}

type KeyQueryStep struct {
	*defaultQueryStep
	Term  string
	IsNum bool
	Num   int
}

var _ QueryStep = (*KeyQueryStep)(nil)

func (qs *KeyQueryStep) String() string {
	return "[" + qs.Term + "]"
}

func (qs *KeyQueryStep) Kind() QueryStepKind {
	return KeyQueryStepKind
}

//func (qs *KeyQueryStep) GetElement(val protoreflect.Value) (protoreflect.Value, bool) {
//	if list.IsValid() && qs.Num < list.Len() {
//		return list.Get(qs.index), true
//	}
//	return protoreflect.Value{}, false
//}
