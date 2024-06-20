package protoquery

import "fmt"

type Type uint8

const (
	TypeBool Type = iota
	TypeString
	TypeNumber
)

type Expression interface {
	Eval() (any, error)
	Type() Type
	String() string
}

type Literal struct {
	value any
	typ   Type
}

func NewLiteral(value any, typ Type) *Literal {
	return &Literal{
		value: value,
		typ:   typ,
	}
}

var _ Expression = (*Literal)(nil)

func (l *Literal) Eval() (any, error) {
	return l.value, nil
}

func (l *Literal) Type() Type {
	return l.typ
}

func (l *Literal) String() string {
	return fmt.Sprintf("%v", l.value)
}

type Operator uint8

const (
	OpPlus Operator = iota
	OpMinus
	OpMul
	OpDiv
	OpEq
	OpNe
	OpLt
	OpLe
	OpGt
	OpGe
	OpAnd
	OpOr
	OpNot
)

var OpToStr = map[Operator]string{
	OpPlus:  "+",
	OpMinus: "-",
	OpMul:   "*",
	OpDiv:   "/",
	OpEq:    "==",
	OpNe:    "!=",
	OpLt:    "<",
	OpLe:    "<=",
	OpGt:    ">",
	OpGe:    ">=",
	OpAnd:   "&&",
	OpOr:    "||",
	OpNot:   "!",
}

type BinaryExpression struct {
	left, right Expression
	op          Operator
}

func NewBinaryExpression(left Expression, op Operator, right Expression) *BinaryExpression {
	return &BinaryExpression{
		left:  left,
		right: right,
		op:    op,
	}
}

var _ Expression = (*BinaryExpression)(nil)

func (b *BinaryExpression) Eval() (any, error) {
	lt, rt := b.left.Type(), b.right.Type()
	if lt != rt {
		return nil, fmt.Errorf("Type mismatch(%v Vs %v)", lt, rt)
	}
	switch b.op {
	case OpPlus:
		return plus(b.left, b.right)
	default:
		return nil, fmt.Errorf("Invalid operator %v", b.op)
	}
}

func (b *BinaryExpression) Type() Type {
	return b.left.Type()
}

func (b *BinaryExpression) String() string {
	return fmt.Sprintf("%v %v %v", b.left, OpToStr[b.op], b.right)
}

func plus(a, b Expression) (any, error) {
	av, err := a.Eval()
	if err != nil {
		return nil, err
	}
	bv, err := b.Eval()
	if err != nil {
		return nil, err
	}

	switch a.Type() {
	case TypeNumber:
		return av.(int64) + bv.(int64), nil
	case TypeString:
		return av.(string) + bv.(string), nil
	default:
		return nil, fmt.Errorf("Invalid type %v for + operator", a.Type())
	}
}
