package protoquery

import (
	"strings"
)

type QueryStepKind int

const (
	SelfQueryStepKind QueryStepKind = iota
	NodeQueryStepKind
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

type KeyQueryStep struct {
	*defaultQueryStep
	expr Expression
}

var _ QueryStep = (*KeyQueryStep)(nil)

func (qs *KeyQueryStep) String() string {
	return "[" + qs.expr.String() + "]"
}

func (qs *KeyQueryStep) Kind() QueryStepKind {
	return KeyQueryStepKind
}
