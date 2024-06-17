package protoquery

import (
	"cmp"
	"fmt"
	"strconv"
	"strings"

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
)

type Predicate interface {
	Match(protoreflect.Value) bool
	String() string
}

type AttrPredicate struct {
	Name  string
	Value string
	Cmp   AttrCmp
}

var _ Predicate = (*AttrPredicate)(nil)

// cmp.Ordered requires golang 1.21+.
// see https://go.dev/blog/comparable for more details.
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

func (ap *AttrPredicate) Match(val protoreflect.Value) bool {
	if !val.IsValid() {
		return false
	}

	msg := val.Message()
	field := msg.Descriptor().Fields().ByName(protoreflect.Name(ap.Name))
	if field == nil {
		return false
	}
	switch ap.Cmp {
	case AttrCmpExist:
		return msg.Has(field)
	case AttrCmpEq, AttrCmpNe, AttrCmpGt, AttrCmpLt, AttrCmpGe, AttrCmpLe:
		val := msg.Get(field)
		switch field.Kind() {
		case protoreflect.StringKind:
			return compare(val.String(), ap.Value, ap.Cmp)
		case protoreflect.FloatKind:
			f, err := strconv.ParseFloat(ap.Value, 64)
			if err != nil {
				return false
			}
			return compare(val.Float(), f, ap.Cmp)
		case protoreflect.Int32Kind, protoreflect.Sint32Kind,
			protoreflect.Sfixed32Kind, protoreflect.Int64Kind,
			protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
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
	s.WriteString(ap.Name)
	if ap.Cmp != AttrCmpExist {
		s.WriteString(fmt.Sprintf("%s%q", cmpToStr[ap.Cmp], ap.Value))
	}
	return s.String()
}

// TODO(osdrv): I believe this should be gone. It is only used as an intermediate value
// in the Compiler. I need a better way to discriminate between index and field predicates.
type IndexPredicate struct {
	Index int
}

var _ Predicate = (*IndexPredicate)(nil)

func (ip *IndexPredicate) Match(protoreflect.Value) bool {
	// stub method
	return false
}

func (ip *IndexPredicate) String() string {
	return fmt.Sprintf("[%d]", ip.Index)
}
