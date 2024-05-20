package protoquery

import (
	"fmt"
	"strings"
)

type AttrCmp int

const (
	AttrCmpExist AttrCmp = iota
	AttrCmpEq
	AttrCmpNe
	AttrCmpGt
	AttrCmpLt
	AttrCmpGe
	AttrCmpLe
)

var cmpToStr = map[AttrCmp]string{
	AttrCmpExist: "",
	AttrCmpEq:    "=",
	AttrCmpNe:    "!=",
	AttrCmpGt:    ">",
	AttrCmpLt:    "<",
	AttrCmpGe:    ">=",
	AttrCmpLe:    "<=",
}

type QueryStepKind int

const (
	SelfQueryStepKind QueryStepKind = iota
	NodeQueryStepKind
	AttrFilterQueryStepKind
	IndexQueryStepKind
	RootQueryStepKind
	RecursiveDescentQueryStepKind
	ParentQueryStepKind
)

type AttrPredicate struct {
	Name  string
	Value string
	Cmp   AttrCmp
}

type Query []QueryStep

func (q Query) String() string {
	var s strings.Builder
	for _, step := range q {
		s.WriteString(step.Name())
	}
	return s.String()
}

type QueryStep interface {
	Name() string
	Kind() QueryStepKind
	Predicate() *AttrPredicate
}

type SelfQueryStep struct {
	predicate *AttrPredicate
}

var _ QueryStep = (*SelfQueryStep)(nil)

func (qs *SelfQueryStep) Name() string {
	return "."
}

func (qs *SelfQueryStep) Kind() QueryStepKind {
	return SelfQueryStepKind
}

func (qs *SelfQueryStep) Predicate() *AttrPredicate {
	return qs.predicate
}

type NodeQueryStep struct {
	name string
}

var _ QueryStep = (*NodeQueryStep)(nil)

func (qs *NodeQueryStep) Name() string {
	return qs.name
}

func (qs *NodeQueryStep) Kind() QueryStepKind {
	return NodeQueryStepKind
}

func (qs *NodeQueryStep) Predicate() *AttrPredicate {
	return nil
}

type AttrFilterStep struct {
	predicate *AttrPredicate
}

var _ QueryStep = (*AttrFilterStep)(nil)

func (qs *AttrFilterStep) Name() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("[@%s", qs.predicate.Name))
	if qs.predicate.Cmp != AttrCmpExist {
		s.WriteString(fmt.Sprintf("%s%s", cmpToStr[qs.predicate.Cmp], qs.predicate.Value))
	}
	s.WriteString("]")
	return s.String()
}

func (qs *AttrFilterStep) Kind() QueryStepKind {
	return AttrFilterQueryStepKind
}

func (qs *AttrFilterStep) Predicate() *AttrPredicate {
	return qs.predicate
}

type IndexQueryStep struct {
	index int
}

var _ QueryStep = (*IndexQueryStep)(nil)

func (qs *IndexQueryStep) Name() string {
	return fmt.Sprintf("[%d]", qs.index)
}

func (qs *IndexQueryStep) Kind() QueryStepKind {
	return IndexQueryStepKind
}

func (qs *IndexQueryStep) Predicate() *AttrPredicate {
	return nil
}

type RootQueryStep struct {
	predicate *AttrPredicate
}

var _ QueryStep = (*RootQueryStep)(nil)

func (qs *RootQueryStep) Name() string {
	return "/"
}

func (qs *RootQueryStep) Kind() QueryStepKind {
	return RootQueryStepKind
}

func (qs *RootQueryStep) Predicate() *AttrPredicate {
	return qs.predicate
}

type RecursiveDescentQueryStep struct {
}

var _ QueryStep = (*RecursiveDescentQueryStep)(nil)

func (qs *RecursiveDescentQueryStep) Name() string {
	return "//"
}

func (qs *RecursiveDescentQueryStep) Kind() QueryStepKind {
	return RecursiveDescentQueryStepKind
}

func (qs *RecursiveDescentQueryStep) Predicate() *AttrPredicate {
	return nil
}

type ParentQueryStep struct {
	predicate *AttrPredicate
}

var _ QueryStep = (*ParentQueryStep)(nil)

func (qs *ParentQueryStep) Name() string {
	return ".."
}

func (qs *ParentQueryStep) Kind() QueryStepKind {
	return ParentQueryStepKind
}

func (qs *ParentQueryStep) Predicate() *AttrPredicate {
	return qs.predicate
}
