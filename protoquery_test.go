package protoquery

import (
	"reflect"
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
			if !reflect.DeepEqual(res, tt.want) {
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
			if !reflect.DeepEqual(res, tt.want) {
				t.Errorf("Compile() = %+v, want %+v", res, tt.want)
			}
		})
	}
}
