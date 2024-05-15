package protoquery

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protopath"
)

type ProtoQuery struct {
	path protopath.Path
}

func Compile(query string) (*ProtoQuery, error) {
	return &ProtoQuery{}, nil
}

func (pq *ProtoQuery) Find(root proto.Message) (proto.Message, error) {
	panic("not implemented")
}

func (pq *ProtoQuery) FindAll(root proto.Message) []proto.Message {
	return nil
}
