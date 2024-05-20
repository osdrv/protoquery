package protoquery

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
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

var (
	cmpToStr = map[AttrCmp]string{
		AttrCmpExist: "",
		AttrCmpEq:    "=",
		AttrCmpNe:    "!=",
		AttrCmpGt:    ">",
		AttrCmpLt:    "<",
		AttrCmpGe:    ">=",
		AttrCmpLe:    "<=",
	}

	allMatchPredicate = &AllMatchPredicate{}
)

type Predicate interface {
	IsMatch(int, proto.Message) bool
	String() string
}

type AttrPredicate struct {
	Name  string
	Value string
	Cmp   AttrCmp
}

var _ Predicate = (*AttrPredicate)(nil)

func (ap *AttrPredicate) IsMatch(index int, msg proto.Message) bool {
	// TODO(osdrv): implement me
	return true
}

func (ap *AttrPredicate) String() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("[@%s", ap.Name))
	if ap.Cmp != AttrCmpExist {
		s.WriteString(fmt.Sprintf("%s%s", cmpToStr[ap.Cmp], ap.Value))
	}
	s.WriteString("]")
	return s.String()
}

type IndexPredicate struct {
	Index int
}

var _ Predicate = (*IndexPredicate)(nil)

func (ip *IndexPredicate) IsMatch(index int, msg proto.Message) bool {
	return index == ip.Index
}

func (ip *IndexPredicate) String() string {
	return fmt.Sprintf("[%d]", ip.Index)
}

type AllMatchPredicate struct{}

var _ Predicate = (*AllMatchPredicate)(nil)

func (ap *AllMatchPredicate) IsMatch(index int, msg proto.Message) bool {
	return true
}

func (ap *AllMatchPredicate) String() string {
	return "[*]"
}

type AndPredicate struct {
	predicates []Predicate
}

var _ Predicate = (*AndPredicate)(nil)

func (ap *AndPredicate) IsMatch(index int, msg proto.Message) bool {
	for _, p := range ap.predicates {
		if !p.IsMatch(index, msg) {
			return false
		}
	}
	return true
}

func (ap *AndPredicate) String() string {
	var s strings.Builder
	s.WriteString("[")
	for i, p := range ap.predicates {
		if i > 0 {
			s.WriteString(" and ")
		}
		s.WriteString(p.String())
	}
	s.WriteString("]")
	return s.String()
}

func (ap *AndPredicate) And(other Predicate) {
	ap.predicates = append(ap.predicates, other)
}

type OrPredicate struct {
	predicates []Predicate
}

var _ Predicate = (*OrPredicate)(nil)

func (op *OrPredicate) IsMatch(index int, msg proto.Message) bool {
	for _, p := range op.predicates {
		if p.IsMatch(index, msg) {
			return true
		}
	}
	return false
}

func (op *OrPredicate) String() string {
	var s strings.Builder
	s.WriteString("[")
	for i, p := range op.predicates {
		if i > 0 {
			s.WriteString(" or ")
		}
		s.WriteString(p.String())
	}
	return s.String()
}

func (op *OrPredicate) Or(other Predicate) {
	op.predicates = append(op.predicates, other)
}
