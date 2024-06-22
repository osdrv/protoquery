package protoquery

import (
	"fmt"
	"strings"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
)

type Type uint8

const (
	TypeBool Type = iota
	TypeString
	TypeNumber
)

type Operator uint8

const (
	_ Operator = iota
	OpPlus
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
	OpEq:    "=",
	OpNe:    "!=",
	OpLt:    "<",
	OpLe:    "<=",
	OpGt:    ">",
	OpGe:    ">=",
	OpAnd:   "&&",
	OpOr:    "||",
	OpNot:   "!",
}

type BuildinContext map[string]any

type Buildin struct {
	body func(ctx BuildinContext, args []Expression) (any, error)
	typ  Type
}

var (
	buildinTypes = map[string]Type{
		"last":     TypeNumber,
		"length":   TypeNumber,
		"position": TypeNumber,
	}
)

type EvalContext struct {
	This any
	kv   map[string]any
}

func NewEvalContext(this any) *EvalContext {
	return &EvalContext{
		This: this,
		kv:   make(map[string]any),
	}
}

type Expression interface {
	Eval(*EvalContext) (any, error)
	Type(*EvalContext) Type
	String() string
}

type LiteralExpr struct {
	value any
	typ   Type
}

func NewLiteralExpr(value any, typ Type) *LiteralExpr {
	return &LiteralExpr{
		value: value,
		typ:   typ,
	}
}

var _ Expression = (*LiteralExpr)(nil)

func (l *LiteralExpr) Eval(*EvalContext) (any, error) {
	return l.value, nil
}

func (l *LiteralExpr) Type(*EvalContext) Type {
	return l.typ
}

func (l *LiteralExpr) String() string {
	return fmt.Sprintf("%v", l.value)
}

type PropertyExpr struct {
	name string
}

var _ Expression = (*PropertyExpr)(nil)

func NewPropertyExpr(name string) *PropertyExpr {
	return &PropertyExpr{
		name: name,
	}
}

var (
	PropNotSet = fmt.Errorf("Property not set")
)

func (p *PropertyExpr) Eval(ctx *EvalContext) (any, error) {
	// TODO(osdrv): implement wildcard
	msg, ok := ctx.This.(protoreflect.Message)
	if !ok {
		return nil, fmt.Errorf("Invalid context %T, want: protoreflect.Message", ctx)
	}
	fd := msg.Descriptor().Fields().ByName(protoreflect.Name(p.name))
	if msg.Has(fd) {
		return msg.Get(fd).Interface(), nil
	}
	return nil, PropNotSet
}

func (p *PropertyExpr) Type(ctx *EvalContext) Type {
	// TODO(osdrv): it can be of any type!
	return TypeString
}

func (p *PropertyExpr) String() string {
	return fmt.Sprintf("@%v", p.name)
}

type FunctionCallExpr struct {
	handle string
	args   []Expression
	typ    Type
}

var _ Expression = (*FunctionCallExpr)(nil)

func NewFunctionCallExpr(handle string, args []Expression) (*FunctionCallExpr, error) {
	typ, ok := buildinTypes[handle]
	if !ok {
		return nil, fmt.Errorf("Unknown function invocation: %v", handle)
	}
	return &FunctionCallExpr{
		handle: handle,
		args:   args,
		typ:    typ,
	}, nil
}

func (f *FunctionCallExpr) Eval(ctx *EvalContext) (any, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (f *FunctionCallExpr) Type(*EvalContext) Type {
	return buildinTypes[f.handle]
}

func (f *FunctionCallExpr) String() string {
	var b strings.Builder
	b.WriteString(f.handle)
	b.WriteString("(")
	for i, arg := range f.args {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(arg.String())
	}
	b.WriteString(")")
	return b.String()
}

type UnaryExpr struct {
	expr Expression
	op   Operator
}

var _ Expression = (*UnaryExpr)(nil)

func NewUnaryExpr(op Operator, expr Expression) (*UnaryExpr, error) {
	return &UnaryExpr{
		expr: expr,
		op:   op,
	}, nil
}

func (u *UnaryExpr) Eval(ctx *EvalContext) (any, error) {
	switch u.op {
	case OpMinus, OpPlus:
		var f int64 = 1
		if u.op == OpMinus {
			f = -1
		}
		if u.expr.Type(ctx) != TypeNumber {
			return nil, fmt.Errorf("Invalid type %v for - operator", u.expr.Type(ctx))
		}
		v, err := u.expr.Eval(ctx)
		if err != nil {
			return nil, err
		}
		return f * v.(int64), nil
	case OpNot:
		if u.expr.Type(ctx) != TypeBool {
			return nil, fmt.Errorf("Invalid type %v for ! operator", u.expr.Type(ctx))
		}
		v, err := u.expr.Eval(ctx)
		if err != nil {
			return nil, err
		}
		return !v.(bool), nil
	default:
		return nil, fmt.Errorf("Invalid operator %v", u.op)
	}
}

func (u *UnaryExpr) Type(ctx *EvalContext) Type {
	switch u.op {
	case OpMinus, OpPlus:
		return TypeNumber
	case OpNot:
		return TypeBool
	default:
		panic("Invalid operator")
	}
}

func (u *UnaryExpr) String() string {
	return fmt.Sprintf("%v %v", OpToStr[u.op], u.expr)
}

type BinaryExpr struct {
	left, right Expression
	op          Operator
}

func NewBinaryExpression(left Expression, op Operator, right Expression) (*BinaryExpr, error) {
	return &BinaryExpr{
		left:  left,
		right: right,
		op:    op,
	}, nil
}

var _ Expression = (*BinaryExpr)(nil)

func (b *BinaryExpr) Eval(ctx *EvalContext) (any, error) {
	lt, rt := b.left.Type(ctx), b.right.Type(ctx)
	if lt != rt {
		return nil, fmt.Errorf("Type mismatch(%v Vs %v)", lt, rt)
	}
	switch b.op {
	case OpPlus:
		switch lt {
		case TypeNumber:
			return numericBinEval(ctx, b.left, b.right, b.op)
		case TypeString:
			return stringBinEval(ctx, b.left, b.right, b.op)
		case TypeBool:
			return boolBinEval(ctx, b.left, b.right, b.op)
		default:
			return nil, fmt.Errorf("Invalid type %v for + operator", lt)
		}
	case OpMinus, OpDiv, OpMul:
		return numericBinEval(ctx, b.left, b.right, b.op)
	case OpAnd, OpOr:
		return boolBinEval(ctx, b.left, b.right, b.op)
	default:
		return nil, fmt.Errorf("Invalid operator %v", b.op)
	}
}

func (b *BinaryExpr) Type(ctx *EvalContext) Type {
	return b.left.Type(ctx)
}

func (b *BinaryExpr) String() string {
	return fmt.Sprintf("%v %v %v", b.left, OpToStr[b.op], b.right)
}

func numericBinEval(ctx *EvalContext, a, b Expression, op Operator) (any, error) {
	if a.Type(ctx) != TypeNumber {
		return nil, fmt.Errorf("Invalid type %v for %v operator", a.Type(ctx), op)
	}
	av, err := a.Eval(ctx)
	if err != nil {
		return nil, err
	}
	bv, err := b.Eval(ctx)
	if err != nil {
		return nil, err
	}
	switch op {
	case OpPlus:
		return av.(int64) + bv.(int64), nil
	case OpMinus:
		return av.(int64) - bv.(int64), nil
	case OpMul:
		return av.(int64) * bv.(int64), nil
	case OpDiv:
		return av.(int64) / bv.(int64), nil
	default:
		return nil, fmt.Errorf("Invalid operator %v", op)
	}
}

func stringBinEval(ctx *EvalContext, a, b Expression, op Operator) (any, error) {
	if a.Type(ctx) != TypeString {
		return nil, fmt.Errorf("Invalid type %v for %v operator", a.Type(ctx), op)
	}
	av, err := a.Eval(ctx)
	if err != nil {
		return nil, err
	}
	bv, err := b.Eval(ctx)
	if err != nil {
		return nil, err
	}
	switch op {
	case OpPlus:
		return av.(string) + bv.(string), nil
	default:
		return nil, fmt.Errorf("Invalid operator %v", op)
	}
}

func boolBinEval(ctx *EvalContext, a, b Expression, op Operator) (any, error) {
	if a.Type(ctx) != TypeBool {
		return nil, fmt.Errorf("Invalid type %v for %v operator", a.Type(ctx), op)
	}
	av, err := a.Eval(ctx)
	if err != nil {
		return nil, err
	}
	bv, err := b.Eval(ctx)
	if err != nil {
		return nil, err
	}
	switch op {
	case OpAnd:
		return av.(bool) && bv.(bool), nil
	case OpOr:
		return av.(bool) || bv.(bool), nil
	default:
		return nil, fmt.Errorf("Invalid operator %v", op)
	}
}
