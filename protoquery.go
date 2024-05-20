package protoquery

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protopath"
)

type ProtoQuery struct {
	query Query
}

func Compile(q string) (*ProtoQuery, error) {
	tokens, err := tokenizeXPathQuery(q)
	if err != nil {
		return nil, err
	}
	query, err := compileQuery(tokens)
	if err != nil {
		return nil, err
	}
	return &ProtoQuery{
		query: query,
	}, nil
}

func (pq *ProtoQuery) Find(root proto.Message) (proto.Message, error) {
	panic("not implemented")
}

func (pq *ProtoQuery) FindAll(msg proto.Message) []proto.Message {
	res := []proto.Message{}
	root := protopath.Root(msg.ProtoReflect().Descriptor())
	fmt.Printf("root=%+v", root)

	return res
}
