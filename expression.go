package protoquery

import (
	"fmt"
	"strings"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
)

type Type uint8

const (
	TypeUnknown Type = iota
	TypeBool
	TypeString
	TypeInt
	TypeFloat
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

type Builtin struct {
	body func(ctx EvalContext, args []Expression) (any, error)
	typ  Type
}

func (b *Builtin) Call(ctx EvalContext, args []Expression) (any, error) {
	return b.body(ctx, args)
}

var (
	builtins = map[string]Builtin{
		"length": {
			body: func(ctx EvalContext, args []Expression) (any, error) {
				list, ok := ctx.This().(protoreflect.List)
				if !ok {
					return nil, fmt.Errorf("length() is only supported for repeated fields")
				}
				return list.Len(), nil
			},
			typ: TypeInt,
		},
		"position": {
			body: func(ctx EvalContext, args []Expression) (any, error) {
				ictx, ok := ctx.(IndexedEvalContext)
				if !ok {
					return nil, fmt.Errorf("position() expects an indexed context")
				}
				return ictx.Index(), nil
			},
			typ: TypeInt,
		},
	}
)

type Expression interface {
	Eval(EvalContext) (any, error)
	Type(EvalContext) (Type, error)
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

func (l *LiteralExpr) Eval(EvalContext) (any, error) {
	switch l.typ {
	case TypeBool:
		return l.value.(bool), nil
	case TypeString:
		return l.value.(string), nil
	case TypeInt:
		intv, err := toInt64(l.value)
		if err != nil {
			return nil, err
		}
		return intv, nil
	case TypeFloat:
		floatv, err := toFloat64(l.value)
		if err != nil {
			return nil, err
		}
		return floatv, nil
	default:
		return nil, fmt.Errorf("Unknown type %v", l.typ)
	}
}

func (l *LiteralExpr) Type(EvalContext) (Type, error) {
	return l.typ, nil
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

func (p *PropertyExpr) Eval(ctx EvalContext) (any, error) {
	// TODO(osdrv): implement wildcard
	msg, ok := ctx.This().(protoreflect.Message)
	if !ok {
		return nil, fmt.Errorf("Invalid list value %T, want: protoreflect.Message", ctx.This())
	}
	fd := msg.Descriptor().Fields().ByName(protoreflect.Name(p.name))
	if fd != nil {
		if ctx.Options().EnforceBool {
			return msg.Has(fd), nil
		}
		if msg.Has(fd) {
			return msg.Get(fd).Interface(), nil
		} else if ctx.Options().UseDefault {
			return fd.Default().Interface(), nil
		}
	}
	return nil, PropNotSet
}

func (p *PropertyExpr) Type(ctx EvalContext) (Type, error) {
	if ctx.Options().EnforceBool {
		return TypeBool, nil
	}
	// TODO(osdrv): in the future we might pass primitive types directly
	// to support `.` (this) operator.
	msg, ok := ctx.This().(protoreflect.Message)
	if !ok {
		return TypeUnknown, fmt.Errorf("Invalid list value %T, want: protoreflect.Message", ctx.This())
	}
	fd, ok := findFieldByName(msg.Interface(), p.name)
	if !ok {
		return TypeUnknown, fmt.Errorf("Field %v not found", p.name)
	}
	switch fd.Kind() {
	case protoreflect.BoolKind:
		return TypeBool, nil
	case protoreflect.StringKind:
		return TypeString, nil
	case protoreflect.Int32Kind, protoreflect.Int64Kind,
		protoreflect.Uint32Kind, protoreflect.Uint64Kind:
		return TypeInt, nil
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return TypeFloat, nil
	}
	return TypeUnknown, fmt.Errorf("Unknown field type %v", fd.Kind())
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
	builtin, ok := builtins[handle]
	if !ok {
		return nil, fmt.Errorf("Unknown function invocation: %v", handle)
	}
	return &FunctionCallExpr{
		handle: handle,
		args:   args,
		typ:    builtin.typ,
	}, nil
}

func (f *FunctionCallExpr) Eval(ctx EvalContext) (any, error) {
	builtin, ok := builtins[f.handle]
	if !ok {
		return nil, fmt.Errorf("Unknown function %v", f.handle)
	}
	return builtin.Call(ctx, f.args)
}

func (f *FunctionCallExpr) Type(EvalContext) (Type, error) {
	return builtins[f.handle].typ, nil
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

func (u *UnaryExpr) Eval(ctx EvalContext) (any, error) {
	switch u.op {
	case OpMinus, OpPlus:
		var f int64 = 1
		if u.op == OpMinus {
			f = -1
		}
		typ, err := u.expr.Type(ctx)
		if err != nil {
			return nil, err
		}
		if typ != TypeInt {
			return nil, fmt.Errorf("Invalid type %v for - operator", typ)
		}
		v, err := u.expr.Eval(ctx)
		if err != nil {
			return nil, err
		}
		intv, err := toInt64(v)
		if err != nil {
			return nil, err
		}
		return f * intv, nil
	case OpNot:
		typ, err := u.expr.Type(ctx)
		if err != nil {
			return nil, err
		}
		if typ != TypeBool {
			return nil, fmt.Errorf("Invalid type %v for ! operator", typ)
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

func (u *UnaryExpr) Type(ctx EvalContext) (Type, error) {
	switch u.op {
	case OpMinus, OpPlus:
		return TypeInt, nil
	case OpNot:
		return TypeBool, nil
	default:
		return TypeUnknown, fmt.Errorf("Invalid operator %v", u.op)
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

func typesCompatible(a, b Type) bool {
	if a == b {
		return true
	}
	// Keep types sorted to avoid duplicate statements here.
	if a > b {
		a, b = b, a
	}
	if a == TypeInt && b == TypeFloat {
		return true
	}

	return false
}

func (b *BinaryExpr) Eval(ctx EvalContext) (any, error) {
	ltyp, lerr := b.left.Type(ctx)
	if lerr != nil {
		return nil, lerr
	}
	rtyp, rerr := b.right.Type(ctx)
	if rerr != nil {
		return nil, rerr
	}
	if !typesCompatible(ltyp, rtyp) {
		return nil, fmt.Errorf("Type mismatch(%v Vs %v)", ltyp, rtyp)
	}
	switch b.op {
	case OpEq, OpNe:
		switch ltyp {
		case TypeInt, TypeFloat:
			return numericBinEval(ctx.Copy(WithUseDefault(true)), b.left, b.right, b.op)
		case TypeString:
			return stringBinEval(ctx.Copy(WithUseDefault(true)), b.left, b.right, b.op)
		case TypeBool:
			return boolBinEval(ctx.Copy(WithUseDefault(true)), b.left, b.right, b.op)
		default:
			return nil, fmt.Errorf("Invalid type %v for + operator", ltyp)
		}
	case OpPlus, OpLt, OpLe, OpGt, OpGe:
		switch ltyp {
		case TypeInt, TypeFloat:
			return numericBinEval(ctx, b.left, b.right, b.op)
		case TypeString:
			return stringBinEval(ctx, b.left, b.right, b.op)
		default:
			return nil, fmt.Errorf("Invalid type %v for %v operator", ltyp, b.op)
		}
	case OpMinus, OpDiv, OpMul:
		return numericBinEval(ctx, b.left, b.right, b.op)
	case OpAnd, OpOr:
		return boolBinEval(ctx, b.left, b.right, b.op)
	default:
		return nil, fmt.Errorf("Invalid operator %v", b.op)
	}
}

func (b *BinaryExpr) Type(ctx EvalContext) (Type, error) {
	switch b.op {
	case OpEq, OpNe, OpLt, OpLe, OpGt, OpGe, OpAnd, OpOr:
		return TypeBool, nil
	default:

		return b.left.Type(ctx)
	}
}

func (b *BinaryExpr) String() string {
	return fmt.Sprintf("%v %v %v", b.left, OpToStr[b.op], b.right)
}

func numericBinEval(ctx EvalContext, a, b Expression, op Operator) (any, error) {
	atyp, aerr := a.Type(ctx)
	if aerr != nil {
		return nil, aerr
	}
	if atyp != TypeInt && atyp != TypeFloat {
		return nil, fmt.Errorf("Invalid type %v for %v operator", atyp, op)
	}
	btyp, berr := b.Type(ctx)
	if berr != nil {
		return nil, berr
	}
	if btyp != TypeInt && btyp != TypeFloat {
		return nil, fmt.Errorf("Invalid type %v for %v operator", btyp, op)
	}
	av, err := a.Eval(ctx)
	if err != nil {
		return nil, err
	}
	bv, err := b.Eval(ctx)
	if err != nil {
		return nil, err
	}
	// Coallesce types to float64 if they are both numeric but do not match.
	if atyp != btyp {
		if atyp == TypeInt {
			av = float64(av.(int64))
			atyp = TypeFloat
		}
		if btyp == TypeInt {
			bv = float64(bv.(int64))
			btyp = TypeFloat
		}
	}
	if atyp == TypeInt {
		ai, aerr := toInt64(av)
		if aerr != nil {
			return nil, aerr
		}
		bi, berr := toInt64(bv)
		if berr != nil {
			return nil, berr
		}
		switch op {
		case OpPlus:
			return ai + bi, nil
		case OpMinus:
			return ai - bi, nil
		case OpMul:
			return ai * bi, nil
		case OpDiv:
			return ai / bi, nil
		case OpEq:
			return ai == bi, nil
		case OpNe:
			return ai != bi, nil
		case OpLt:
			return ai < bi, nil
		case OpLe:
			return ai <= bi, nil
		case OpGt:
			return ai > bi, nil
		case OpGe:
			return ai >= bi, nil
		default:
			return nil, fmt.Errorf("Invalid operator %v", op)
		}
	} else if atyp == TypeFloat {
		af, aerr := toFloat64(av)
		if aerr != nil {
			return nil, aerr
		}
		bf, berr := toFloat64(bv)
		if berr != nil {
			return nil, berr
		}
		switch op {
		case OpPlus:
			return af + bf, nil
		case OpMinus:
			return af - bf, nil
		case OpMul:
			return af * bf, nil
		case OpDiv:
			return af / bf, nil
		case OpEq:
			return af == bf, nil
		case OpNe:
			return af != bf, nil
		case OpLt:
			return af < bf, nil
		case OpLe:
			return af <= bf, nil
		case OpGt:
			return af > bf, nil
		case OpGe:
			return af >= bf, nil
		default:
			return nil, fmt.Errorf("Invalid operator %v", op)
		}
	}
	panic("should not happen")
}

func stringBinEval(ctx EvalContext, a, b Expression, op Operator) (any, error) {
	atyp, aerr := a.Type(ctx)
	if aerr != nil {
		return nil, aerr
	}
	if atyp != TypeString {
		return nil, fmt.Errorf("Invalid type %v for %v operator", atyp, op)
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
	case OpEq:
		return av.(string) == bv.(string), nil
	case OpNe:
		return av.(string) != bv.(string), nil
	case OpLt:
		return av.(string) < bv.(string), nil
	case OpLe:
		return av.(string) <= bv.(string), nil
	case OpGt:
		return av.(string) > bv.(string), nil
	case OpGe:
		return av.(string) >= bv.(string), nil
	default:
		return nil, fmt.Errorf("Invalid operator %v", op)
	}
}

func boolBinEval(ctx EvalContext, a, b Expression, op Operator) (any, error) {
	atyp, aerr := a.Type(ctx)
	if aerr != nil {
		return nil, aerr
	}
	if atyp != TypeBool {
		return nil, fmt.Errorf("Invalid type %v for %v operator", atyp, op)
	}
	av, err := a.Eval(ctx)
	if err != nil {
		return nil, err
	}

	// There are cases where evaluation of the right side can be avoided.
	if op == OpAnd && !av.(bool) {
		return false, nil
	} else if op == OpOr && av.(bool) {
		return true, nil
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
	case OpEq:
		return av.(bool) == bv.(bool), nil
	case OpNe:
		return av.(bool) != bv.(bool), nil
	default:
		return nil, fmt.Errorf("Invalid operator %v", op)
	}
}

// isAllProps is a helper function that traverses the expression tree and returns true
// if all the expressions are either properties or binary AND/OR expressions.
func isAllPropertyExprs(e Expression) bool {
	if e == nil {
		return false
	}
	switch e.(type) {
	case *PropertyExpr:
		return true
	case *BinaryExpr:
		be := e.(*BinaryExpr)
		return (be.op == OpAnd || be.op == OpOr) &&
			isAllPropertyExprs(be.left) && isAllPropertyExprs(be.right)
	default:
		return false
	}
}
