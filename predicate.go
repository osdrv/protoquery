package protoquery

import (
	"cmp"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
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

func compare[T cmp.Ordered](a, b T, cmp AttrCmp) bool {
	switch cmp {
	case AttrCmpEq:
		return a == b
	case AttrCmpNe:
		return a != b
	case AttrCmpGt:
		return a > b
	case AttrCmpLt:
		return a < b
	case AttrCmpGe:
		return a >= b
	case AttrCmpLe:
		return a <= b
	default:
		panic("comparison operator not implemented")
	}
}

func (ap *AttrPredicate) IsMatch(index int, msg proto.Message) bool {
	if msg == nil {
		return false
	}
	field := msg.ProtoReflect().Descriptor().Fields().ByName(protoreflect.Name(ap.Name))
	if field == nil {
		return false
	}
	switch ap.Cmp {
	case AttrCmpExist:
		return msg.ProtoReflect().Has(field)
	case AttrCmpEq, AttrCmpNe, AttrCmpGt, AttrCmpLt, AttrCmpGe, AttrCmpLe:
		val := msg.ProtoReflect().Get(field)
		switch field.Kind() {
		case protoreflect.StringKind:
			return compare(val.String(), ap.Value, ap.Cmp)
		case protoreflect.FloatKind:
			f, err := strconv.ParseFloat(ap.Value, 64)
			if err != nil {
				return false
			}
			return compare(val.Float(), f, ap.Cmp)
		case protoreflect.Int32Kind,
			protoreflect.Sint32Kind,
			protoreflect.Sfixed32Kind,
			protoreflect.Int64Kind,
			protoreflect.Sint64Kind,
			protoreflect.Sfixed64Kind:
			i, err := strconv.ParseInt(ap.Value, 10, 64)
			if err != nil {
				return false
			}
			return compare(val.Int(), i, ap.Cmp)
		default:
			panic("unsupported match predicate type")
		}
	default:
		panic("not implemented")
	}
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
