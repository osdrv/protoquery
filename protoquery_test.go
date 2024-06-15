package protoquery

import (
	"testing"
)

func TestFindAllAttributeAccess(t *testing.T) {
	store := Bookstore{
		Books: []*Book{
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
		want    []interface{}
		wantErr error
	}{
		{
			name:  "child element attributes",
			query: "/books/book[@author]/author",
			want: []interface{}{
				"Alan A. A. Donovan",
				"Steve Klabnik",
			},
		},
		{
			name:  "chile element with predicate",
			query: "/books/book[@price>35]/price",
			want:  []interface{}{39.99},
		},
		{
			name:  "child element with an empty attribute value",
			query: "/books/book[@title='The Bible']/author",
			want:  []interface{}{""},
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
	store := Bookstore{
		Books: []*Book{
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
		want    []interface{}
		wantErr error
	}{
		{
			name:  "children elements",
			query: "/books/book",
			want: []interface{}{
				store.Books[0],
				store.Books[1],
				store.Books[2],
			},
		},
		{
			name:  "child element by numeric index",
			query: "/books/book[1]",
			want: []interface{}{
				store.Books[1],
			},
		},
		{
			name:  "child element by attribute presence",
			query: "/books/book[@author]",
			want: []interface{}{
				store.Books[0],
				store.Books[1],
			},
		},
		{
			name:  "child element by attribute equality",
			query: "/books/book[@author='Alan A. A. Donovan']",
			want: []interface{}{
				store.Books[0],
			},
		},
		{
			name:  "child element by attribute inequality",
			query: "/books/book[@author!='Alan A. A. Donovan']",
			want: []interface{}{
				store.Books[1],
				store.Books[2],
			},
		},
		{
			name:  "child element with a numeric attribute comparison",
			query: "/books/book[@price>35]",
			want: []interface{}{
				store.Books[1],
			},
		},
		{
			name:  "child element with a wrong name",
			query: "/books/wrong_name[@price>35]",
			want:  []interface{}{},
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
	holder := RepeatedScalarHolder{
		Items: []*RepeatedScalarsItem{
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
		want  []interface{}
	}{
		{
			name:  "return int32 repeated attribute",
			query: "/items/repeated_scalars_item/int32s",
			want:  []interface{}{[]int32{1, 2, 3}},
		},
		{
			name:  "return int64 repeated attribute",
			query: "/items/repeated_scalars_item/int64s",
			want:  []interface{}{[]int64{1, 2, 3}},
		},
		{
			name:  "return uint32 repeated attribute",
			query: "/items/repeated_scalars_item/uint32s",
			want:  []interface{}{[]uint32{1, 2, 3}},
		},
		{
			name:  "return uint64 repeated attribute",
			query: "/items/repeated_scalars_item/uint64s",
			want:  []interface{}{[]uint64{1, 2, 3}},
		},
		{
			name:  "return float repeated attribute",
			query: "/items/repeated_scalars_item/floats",
			want:  []interface{}{[]float32{1.1, 2.2, 3.3}},
		},
		{
			name:  "return string repeated attribute",
			query: "/items/repeated_scalars_item/strings",
			want:  []interface{}{[]string{"a", "b", "c"}},
		},
		{
			name:  "return bools repeated attribute",
			query: "/items/repeated_scalars_item/bools",
			want:  []interface{}{[]bool{true, false}},
		},
		{
			name:  "return bytes repeated attribute",
			query: "/items/repeated_scalars_item/bytes",
			want:  []interface{}{[]byte{1, 2, 3}},
		},
		{
			name:  "return int32 individual attribute",
			query: "/items/repeated_scalars_item/int32s[0]",
			want:  []interface{}{int32(1)},
		},
		{
			name:  "return int64 individual attribute",
			query: "/items/repeated_scalars_item/int64s[0]",
			want:  []interface{}{int64(1)},
		},
		{
			name:  "return uint32 individual attribute",
			query: "/items/repeated_scalars_item/uint32s[0]",
			want:  []interface{}{uint32(1)},
		},
		{
			name:  "return uint64 individual attribute",
			query: "/items/repeated_scalars_item/uint64s[0]",
			want:  []interface{}{uint64(1)},
		},
		{
			name:  "return float individual attribute",
			query: "/items/repeated_scalars_item/floats[0]",
			want:  []interface{}{1.1},
		},
		{
			name:  "return string individual attribute",
			query: "/items/repeated_scalars_item/strings[0]",
			want:  []interface{}{"a"},
		},
		{
			name:  "return bools individual attribute",
			query: "/items/repeated_scalars_item/bools[0]",
			want:  []interface{}{true},
		},
		{
			name:  "return bytes attribute",
			query: "/items/repeated_scalars_item/bytes[0]",
			want:  []interface{}{byte(1)},
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
