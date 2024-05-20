package protoquery

import (
	"fmt"
	"strings"
)

type QueryStepKind int

const (
	SelfQueryStepKind QueryStepKind = iota
	NodeQueryStepKind
	AttrFilterQueryStepKind
	IndexQueryStepKind
	RootQueryStepKind
	RecursiveDescentQueryStepKind
)

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
	Predicate() Predicate
	IntValue() (int, error)
}

type defaultQueryStep struct{}

func (qs *defaultQueryStep) Predicate() Predicate {
	return allMatchPredicate
}

func (qs *defaultQueryStep) IntValue() (int, error) {
	return 0, fmt.Errorf("Not an index query step")
}

type SelfQueryStep struct {
	*defaultQueryStep
}

var _ QueryStep = (*SelfQueryStep)(nil)

func (qs *SelfQueryStep) Name() string {
	return "."
}

func (qs *SelfQueryStep) Kind() QueryStepKind {
	return SelfQueryStepKind
}

type NodeQueryStep struct {
	*defaultQueryStep
	name      string
	predicate Predicate
}

var _ QueryStep = (*NodeQueryStep)(nil)

func (qs *NodeQueryStep) Name() string {
	return qs.name
}

func (qs *NodeQueryStep) Kind() QueryStepKind {
	return NodeQueryStepKind
}

func (qs *NodeQueryStep) Predicate() Predicate {
	if qs.predicate == nil {
		return allMatchPredicate
	}
	return qs.predicate
}

type RootQueryStep struct {
	*defaultQueryStep
}

var _ QueryStep = (*RootQueryStep)(nil)

func (qs *RootQueryStep) Name() string {
	return "/"
}

func (qs *RootQueryStep) Kind() QueryStepKind {
	return RootQueryStepKind
}

type RecursiveDescentQueryStep struct {
	*defaultQueryStep
}

var _ QueryStep = (*RecursiveDescentQueryStep)(nil)

func (qs *RecursiveDescentQueryStep) Name() string {
	return "//"
}

func (qs *RecursiveDescentQueryStep) Kind() QueryStepKind {
	return RecursiveDescentQueryStepKind
}
