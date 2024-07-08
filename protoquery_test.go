package protoquery

import (
	"osdrv/protoquery/proto"
	"testing"
)

func TestFindAllAttributeAccess(t *testing.T) {
	store := proto.Bookstore{
		Books: []*proto.Book{
			{
				Title:  "The Go Programming Language",
				Author: "Alan A. A. Donovan",
				Price:  34.99,
			},
			{
				Title:  "The Rust Programming Language",
				Author: "Steve Klabnik",
				Price:  39.99,
			},
			{
				Title: "The Bible",
				Price: 0.00,
			},
		},
	}

	tests := []struct {
		name    string
		query   string
		want    []any
		wantErr error
	}{
		{
			name:  "child element attributes",
			query: "/books[@author]/author",
			want: []any{
				"Alan A. A. Donovan",
				"Steve Klabnik",
			},
		},
		{
			name:  "child element with predicate",
			query: "/books[@price>35]/price",
			want:  []any{float32(39.99)},
		},
		{
			name:  "child element with an empty attribute value",
			query: "/books[@title='The Bible']/author",
			want:  []any{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pq, err := Compile(tt.query)
			if !errorsSimilar(tt.wantErr, err) {
				t.Errorf("Compile() error = %v, want %v", err, tt.wantErr)
				return
			}
			if err != nil {
				t.Errorf("Compile() error = %v, no error expected", err)
				return
			}
			res := pq.FindAll(&store)
			if !deepEqual(res, tt.want) {
				t.Errorf("Compile() = %+v, want %+v", res, tt.want)
			}
		})
	}
}

func TestFindAllChilrenAccess(t *testing.T) {
	store := proto.Bookstore{
		Books: []*proto.Book{
			{
				Title:  "The Go Programming Language",
				Author: "Alan A. A. Donovan",
				Price:  34.99,
			},
			{
				Title:  "The Rust Programming Language",
				Author: "Steve Klabnik",
				Price:  39.99,
			},
			{
				Title: "The Bible",
				Price: 0.00,
			},
		},
	}

	tests := []struct {
		name    string
		query   string
		want    []any
		wantErr error
	}{
		{
			name:  "children elements",
			query: "/books",
			want: []any{
				store.Books[0],
				store.Books[1],
				store.Books[2],
			},
		},
		{
			name:  "child element by numeric index",
			query: "/books[1]",
			want: []any{
				store.Books[1],
			},
		},
		{
			name:  "child element by attribute presence",
			query: "/books[@author]",
			want: []any{
				store.Books[0],
				store.Books[1],
			},
		},
		{
			name:  "child element by attribute equality",
			query: "/books[@author='Alan A. A. Donovan']",
			want: []any{
				store.Books[0],
			},
		},
		{
			name:  "child element by attribute inequality",
			query: "/books[@author!='Alan A. A. Donovan']",
			want: []any{
				store.Books[1],
				store.Books[2],
			},
		},
		{
			name:  "child element with a numeric attribute comparison",
			query: "/books[@price>35]",
			want: []any{
				store.Books[1],
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pq, err := Compile(tt.query)
			if !errorsSimilar(tt.wantErr, err) {
				t.Errorf("Compile() error = %v, want %v", err, tt.wantErr)
				return
			}
			if err != nil {
				t.Errorf("Compile() error = %v, no error expected", err)
				return
			}
			res := pq.FindAll(&store)
			if !deepEqual(res, tt.want) {
				t.Errorf("Compile() = %+v, want %+v", res, tt.want)
			}
		})
	}
}

func TestFindAllRepeatedScalars(t *testing.T) {
	holder := proto.RepeatedScalarHolder{
		Items: []*proto.RepeatedScalarsItem{
			{
				Int32S:  []int32{1, 2, 3},
				Int64S:  []int64{1, 2, 3},
				Uint32S: []uint32{1, 2, 3},
				Uint64S: []uint64{1, 2, 3},
				Floats:  []float32{1.1, 2.2, 3.3},
				Strings: []string{"a", "b", "c"},
				Bools:   []bool{true, false},
				Bytes:   []byte{1, 2, 3},
			},
		},
	}

	tests := []struct {
		name  string
		query string
		want  []any
	}{
		{
			name:  "return int32 repeated attribute",
			query: "/items/int32s",
			want:  []any{int32(1), int32(2), int32(3)},
		},
		{
			name:  "return int64 repeated attribute",
			query: "/items/int64s",
			want:  []any{int64(1), int64(2), int64(3)},
		},
		{
			name:  "return uint32 repeated attribute",
			query: "/items/uint32s",
			want:  []any{uint32(1), uint32(2), uint32(3)},
		},
		{
			name:  "return uint64 repeated attribute",
			query: "/items/uint64s",
			want:  []any{uint64(1), uint64(2), uint64(3)},
		},
		{
			name:  "return float repeated attribute",
			query: "/items/floats",
			want:  []any{float32(1.1), float32(2.2), float32(3.3)},
		},
		{
			name:  "return string repeated attribute",
			query: "/items/strings",
			want:  []any{"a", "b", "c"},
		},
		{
			name:  "return bools repeated attribute",
			query: "/items/bools",
			want:  []any{true, false},
		},
		{
			name:  "return bytes attribute",
			query: "/items/bytes",
			want:  []any{[]byte{1, 2, 3}},
		},
		{
			name:  "return int32 individual attribute",
			query: "/items/int32s[0]",
			want:  []any{int32(1)},
		},
		{
			name:  "return int64 individual attribute",
			query: "/items/int64s[0]",
			want:  []any{int64(1)},
		},
		{
			name:  "return uint32 individual attribute",
			query: "/items/uint32s[0]",
			want:  []any{uint32(1)},
		},
		{
			name:  "return uint64 individual attribute",
			query: "/items/uint64s[0]",
			want:  []any{uint64(1)},
		},
		{
			name:  "return float individual attribute",
			query: "/items/floats[0]",
			want:  []any{float32(1.1)},
		},
		{
			name:  "return string individual attribute",
			query: "/items/strings[0]",
			want:  []any{"a"},
		},
		{
			name:  "return bools individual attribute",
			query: "/items/bools[0]",
			want:  []any{true},
		},
		{
			name:  "return bytes attribute",
			query: "/items/bytes[0]",
			// This is a corner case. Protoreflect does not support ints with
			// bitness less than 32. So, the bytes are converted to uint32.
			// This must be documented.
			want: []any{uint32(1)},
		},
		{
			name:  "bytes attribute with an out-of-bounds index",
			query: "/items/bytes[100]",
			want:  []any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pq, err := Compile(tt.query)
			if err != nil {
				t.Errorf("Compile() error = %v, no error expected", err)
				return
			}
			res := pq.FindAll(&holder)
			if !deepEqual(res, tt.want) {
				t.Errorf("Compile() = %+v, want %+v", res, tt.want)
			}
		})
	}
}

func TestFindAllMapAccess(t *testing.T) {
	messages := &proto.MessageWithMapHolder{
		MessagesWithMap: []*proto.MessageWithMap{
			{
				StringStringMap: map[string]string{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				},
				StringInnerMap: map[string]*proto.MessageWithMap_InnerMessage{
					"key4": {
						InnerInt:    1,
						InnerString: "string",
						InnerArr:    []int32{1, 2, 3, 4, 5},
					},
				},
			},
			{
				Int32InnerMap: map[int32]*proto.MessageWithMap_InnerMessage{
					int32(123): {InnerString: "map with int32 key"},
				},
			},
			{
				Int64InnerMap: map[int64]*proto.MessageWithMap_InnerMessage{
					int64(123): {InnerString: "map with int64 key"},
				},
			},
			{
				Uint32InnerMap: map[uint32]*proto.MessageWithMap_InnerMessage{
					uint32(123): {InnerString: "map with uint32 key"},
				},
			},
			{
				Uint64InnerMap: map[uint64]*proto.MessageWithMap_InnerMessage{
					uint64(123): {InnerString: "map with uint64 key"},
				},
			},
			{
				Sint32InnerMap: map[int32]*proto.MessageWithMap_InnerMessage{
					int32(123): {InnerString: "map with sint32 key"},
				},
			},
			{
				Sint64InnerMap: map[int64]*proto.MessageWithMap_InnerMessage{
					int64(123): {InnerString: "map with sint64 key"},
				},
			},
			{
				Fixed32InnerMap: map[uint32]*proto.MessageWithMap_InnerMessage{
					uint32(123): {InnerString: "map with fixed32 key"},
				},
			},
			{
				Fixed64InnerMap: map[uint64]*proto.MessageWithMap_InnerMessage{
					uint64(123): {InnerString: "map with fixed64 key"},
				},
			},
			{
				Sfixed32InnerMap: map[int32]*proto.MessageWithMap_InnerMessage{
					int32(123): {InnerString: "map with sfixed32 key"},
				},
			},
			{
				Sfixed64InnerMap: map[int64]*proto.MessageWithMap_InnerMessage{
					int64(123): {InnerString: "map with sfixed64 key"},
				},
			},
			{
				BoolInnerMap: map[bool]*proto.MessageWithMap_InnerMessage{
					true: {InnerString: "map with bool key=true"},
				},
			},
			{
				BoolInnerMap: map[bool]*proto.MessageWithMap_InnerMessage{
					false: {InnerString: "map with bool key=false"},
				},
			},
		},
	}

	tests := []struct {
		name  string
		query string
		want  []any
	}{
		{
			name:  "int32 key lookup",
			query: "/messages_with_map/int32_inner_map[123]/inner_string",
			want:  []any{"map with int32 key"},
		},
		{
			name:  "int64 key lookup",
			query: "/messages_with_map/int64_inner_map[123]/inner_string",
			want:  []any{"map with int64 key"},
		},
		{
			name:  "uint32 key lookup",
			query: "/messages_with_map/uint32_inner_map[123]/inner_string",
			want:  []any{"map with uint32 key"},
		},
		{
			name:  "uint64 key lookup",
			query: "/messages_with_map/uint64_inner_map[123]/inner_string",
			want:  []any{"map with uint64 key"},
		},
		{
			name:  "sint32 key lookup",
			query: "/messages_with_map/sint32_inner_map[123]/inner_string",
			want:  []any{"map with sint32 key"},
		},
		{
			name:  "sint64 key lookup",
			query: "/messages_with_map/sint64_inner_map[123]/inner_string",
			want:  []any{"map with sint64 key"},
		},
		{
			name:  "fixed32 key lookup",
			query: "/messages_with_map/fixed32_inner_map[123]/inner_string",
			want:  []any{"map with fixed32 key"},
		},
		{
			name:  "fixed64 key lookup",
			query: "/messages_with_map/fixed64_inner_map[123]/inner_string",
			want:  []any{"map with fixed64 key"},
		},
		{
			name:  "sfixed32 key lookup",
			query: "/messages_with_map/sfixed32_inner_map[123]/inner_string",
			want:  []any{"map with sfixed32 key"},
		},
		{
			name:  "sfixed64 key lookup",
			query: "/messages_with_map/sfixed64_inner_map[123]/inner_string",
			want:  []any{"map with sfixed64 key"},
		},
		{
			name:  "bool key lookup=true",
			query: "/messages_with_map/bool_inner_map[true]/inner_string",
			want:  []any{"map with bool key=true"},
		},
		{
			name:  "bool key lookup=false",
			query: "/messages_with_map/bool_inner_map[false]/inner_string",
			want:  []any{"map with bool key=false"},
		},
		{
			name:  "string map inner message lookup",
			query: "/messages_with_map/string_inner_map['key4']/inner_int",
			want:  []any{int32(1)},
		},
		{
			name:  "missing key lookup",
			query: "/messages_with_map/string_string_map['key4']",
			want:  []any{},
		},
		{
			name:  "int key lookup on a string-string map",
			query: "/messages_with_map/string_string_map[1]",
			want:  []any{},
		},
		{
			name:  "bool key lookup on a string-string map",
			query: "/messages/with_map/string_string_map[true]",
			want:  []any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pq, err := Compile(tt.query)
			if err != nil {
				t.Errorf("Compile() error = %v, no error expected", err)
				return
			}
			res := pq.FindAll(messages)
			if !deepEqual(res, tt.want) {
				t.Errorf("Compile() = %+v, want %+v", res, tt.want)
			}
		})
	}
}

func TestFindAllListBuiltins(t *testing.T) {
	store := proto.Bookstore{
		Books: []*proto.Book{
			{
				Title:  "The Go Programming Language",
				Author: "Alan A. A. Donovan",
				Price:  34.99,
			},
			{
				Title:  "The Rust Programming Language",
				Author: "Steve Klabnik",
				Price:  39.99,
			},
			{
				Title: "The Bible",
				Price: 0.00,
			},
		},
	}
	tests := []struct {
		name  string
		query string
		want  []any
	}{
		{
			name:  "return the last element of the list",
			query: "/books[length() - 1]",
			want:  []any{store.Books[2]},
		},
		{
			name:  "out-of-bounds index with length",
			query: "/books[length() - 100]",
			want:  []any{},
		},
		{
			name:  "position in a boolean context",
			query: "/books[position() <= 1]",
			want:  []any{store.Books[0], store.Books[1]},
		},
		{
			name:  "position with unexpected type",
			query: "/books[position() > true]",
			want:  []any{},
		},
		{
			name:  "position with unexpected operand",
			query: "/books[position() + 1]",
			want:  []any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pq, err := Compile(tt.query)
			if err != nil {
				t.Errorf("Compile() error = %v, no error expected", err)
				return
			}
			res := pq.FindAll(&store)
			if !deepEqual(res, tt.want) {
				t.Errorf("Compile() = %+v, want %+v", res, tt.want)
			}
		})
	}
}

func TestFindAllEnum(t *testing.T) {
	holder := &proto.MessageWithEnumHolder{
		Messages: []*proto.MessageWithEnum{
			{
				EnumField:   proto.MessageWithEnum_ENUM1,
				StringField: "message with enum1",
			},
			{
				EnumField:   proto.MessageWithEnum_ENUM2,
				StringField: "message with enum2",
			},
			{
				EnumField:   proto.MessageWithEnum_ENUM3,
				StringField: "message with enum3",
			},
		},
	}
	tests := []struct {
		name  string
		query string
		want  []any
	}{
		{
			name:  "single enum selector",
			query: "/messages[@enum_field = 'ENUM1']",
			want:  []any{holder.Messages[0]},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pq, err := Compile(tt.query)
			if err != nil {
				t.Errorf("Compile() error = %v, no error expected", err)
				return
			}
			res := pq.FindAll(holder)
			if !deepEqual(res, tt.want) {
				t.Errorf("Compile() = %+v, want %+v", res, tt.want)
			}
		})
	}
}
