package protoquery

import (
	"fmt"
)

func matchToken(tokens []*Token, ix int, kind TokenKind) bool {
	return ix < len(tokens) && tokens[ix].Kind == kind
}

func matchTokenAny(tokens []*Token, ix int, kinds ...TokenKind) bool {
	for _, kind := range kinds {
		if matchToken(tokens, ix, kind) {
			return true
		}
	}
	return false
}

var (
	tokenToCmp = map[TokenKind]AttrCmp{
		TokenEqual:        AttrCmpEq,
		TokenNotEqual:     AttrCmpNe,
		TokenLess:         AttrCmpLt,
		TokenLessEqual:    AttrCmpLe,
		TokenGreater:      AttrCmpGt,
		TokenGreaterEqual: AttrCmpGe,
	}
)

func compilePredicate(tokens []*Token, ix int) (Predicate, int, error) {
	p := &AttrPredicate{}
	if !matchToken(tokens, ix, TokenAt) {
		return nil, ix, fmt.Errorf("expected @, got %v", tokens[ix].Kind)
	}
	ix++
	if !matchToken(tokens, ix, TokenNode) {
		return nil, ix, fmt.Errorf("expected attribute name, got %v", tokens[ix].Kind)
	}
	p.Name = tokens[ix].Value
	p.Cmp = AttrCmpExist
	ix++

	if matchTokenAny(tokens, ix, TokenEqual, TokenGreater, TokenGreaterEqual,
		TokenNotEqual, TokenLess, TokenLessEqual) {
		p.Cmp = tokenToCmp[tokens[ix].Kind]
		ix++
		if !matchTokenAny(tokens, ix, TokenString, TokenNumber) {
			return nil, ix, fmt.Errorf("expected string or number, got %v", tokens[ix].Kind)
		}
		// TODO(osdrv): are we loosing the type information here?
		p.Value = tokens[ix].Value
		ix++
	}

	return p, ix, nil
}

func compileNodeQueryStep(tokens []*Token, ix int) (*NodeQueryStep, int, error) {
	nqs := &NodeQueryStep{}
	if !matchToken(tokens, ix, TokenNode) {
		return nil, ix, fmt.Errorf("expected node name, got %v", tokens[ix].Kind)
	}
	nqs.name = tokens[ix].Value
	ix++
	return nqs, ix, nil
}

func compileKeyOrAttrFilterQueryStep(tokens []*Token, ix int) (QueryStep, int, error) {
	if !matchToken(tokens, ix, TokenLBracket) {
		return nil, ix, fmt.Errorf("expected [, got %v", tokens[ix].Kind)
	}
	ix++
	var qs QueryStep

	if matchToken(tokens, ix, TokenAt) {
		// Read attribute filter
		var p Predicate
		var err error
		p, ix, err = compilePredicate(tokens, ix)
		if err != nil {
			return nil, ix, err
		}
		qs = &AttrFilterQueryStep{predicate: p}
	} else {
		kqs := &KeyQueryStep{
			Term: tokens[ix].Value,
		}
		// TODO(osdrv): this block of code would have to undergo another round of refactoring when
		// I'll add length() and other expressions. Good enough for now as an intermediate step.
		if num, err := tokens[ix].IntValue(); err == nil {
			kqs.IsNum = true
			kqs.Num = int(num)
		}
		qs = kqs
		ix++
	}

	if !matchToken(tokens, ix, TokenRBracket) {
		return nil, ix, fmt.Errorf("expected ], got %v", tokens[ix].Kind)
	}
	ix++

	return qs, ix, nil
}

func compileQuery(tokens []*Token) (Query, error) {
	var query Query
	ix := 0
	for ix < len(tokens) {
		var err error
		switch tokens[ix].Kind {
		case TokenSlash:
			if len(query) > 0 {
				// Separator between nodes
				ix++
				continue
			}
			ix++
			query = append(query, &RootQueryStep{})
		case TokenSlashSlash:
			query = append(query, &RecursiveDescentQueryStep{})
			ix++
		case TokenNode:
			var qs *NodeQueryStep
			qs, ix, err = compileNodeQueryStep(tokens, ix)
			if err != nil {
				return nil, err
			}
			query = append(query, qs)
		case TokenLBracket:
			var qs QueryStep
			qs, ix, err = compileKeyOrAttrFilterQueryStep(tokens, ix)
			if err != nil {
				return nil, err
			}
			query = append(query, qs)
		default:
			return nil, fmt.Errorf("unexpected token %v %q", tokens[ix].Kind, tokens[ix].Value)
		}
	}
	return query, nil
}

func CompileExpression(tokens []*Token) (Expression, error) {
	// TODO(osdrv): add support for more complex expressions
	expr, _, err := compileLiteralExpression(tokens, 0)
	return expr, err
}

func compileLiteralExpression(tokens []*Token, ix int) (Expression, int, error) {
	if !matchTokenAny(tokens, ix, TokenString, TokenNumber, TokenBool) {
		return nil, ix, fmt.Errorf("expected string or number, got %v", tokens[ix].Kind)
	}
	srcv := tokens[ix].Value
	switch tokens[ix].Kind {
	case TokenString:
		return &Literal{
			value: srcv,
			typ:   TypeString,
		}, ix + 1, nil
	case TokenNumber:
		intv, err := tokens[ix].IntValue()
		if err != nil {
			return nil, ix, err
		}
		return &Literal{
			value: intv,
			typ:   TypeNumber,
		}, ix + 1, nil
	case TokenBool:
		boolv, err := tokens[ix].BoolValue()
		if err != nil {
			return nil, ix, err
		}
		return &Literal{
			value: boolv,
			typ:   TypeBool,
		}, ix + 1, nil
	default:
		return nil, ix, fmt.Errorf("unexpected token %v %q", tokens[ix].Kind, tokens[ix].Value)
	}
}
