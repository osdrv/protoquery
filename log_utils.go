package protoquery

import "fmt"

func debugf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func panicf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}
