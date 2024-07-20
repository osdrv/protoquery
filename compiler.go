package protoquery

import (
	"fmt"
)

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
			qs, ix, err = compileKeyQueryStep(tokens, ix)
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

func compileNodeQueryStep(tokens []*Token, ix int) (*NodeQueryStep, int, error) {
	nqs := &NodeQueryStep{}
	if !matchToken(tokens, ix, TokenNode) {
		return nil, ix, fmt.Errorf("expected node name, got %v", tokens[ix].Value)
	}
	nqs.name = tokens[ix].Value
	ix++
	return nqs, ix, nil
}

func compileKeyQueryStep(tokens []*Token, ix int) (*KeyQueryStep, int, error) {
	var expr Expression
	var err error
	if !matchToken(tokens, ix, TokenLBracket) {
		return nil, ix, fmt.Errorf("expected [, got %v", tokens[ix].Value)
	}
	ix++
	expr, ix, err = parseExpression(tokens, ix, LOWEST)
	if err != nil {
		return nil, ix, err
	}
	if !matchToken(tokens, ix, TokenRBracket) {
		return nil, ix, fmt.Errorf("expected ], got %v", tokens[ix].Value)
	}
	ix++
	return &KeyQueryStep{
		expr: expr,
	}, ix, nil
}
