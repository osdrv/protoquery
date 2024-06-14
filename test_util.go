package protoquery

import (
	reflect "reflect"
	"strings"
)

// errorEqual compares two errors. It returns true if both are nil,
// or if both are not nil and have the same error message.
func errorEqual(err1, err2 error) bool {
	if err1 == nil && err2 == nil {
		return true
	}
	if err1 == nil || err2 == nil {
		return false
	}
	return err1.Error() == err2.Error()
}

// errorSimilar compares two errors. It returns true if both are nil,
// or if both are not nil and the error message of err1 contains the error message of err2.
func errorsSimilar(err1, err2 error) bool {
	if err1 == nil || err2 == nil {
		return err1 == err2
	}
	return strings.Contains(err1.Error(), err2.Error())
}

func equalWithTolerance(a, b, tolerance float64) bool {
	return a-b < tolerance && b-a < tolerance
}

func deepEqual[K any](a, b K) bool {
	valA, valB := reflect.ValueOf(a), reflect.ValueOf(b)
	switch valA.Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return valA.Interface() == valB.Interface()
	case reflect.Float32, reflect.Float64:
		return equalWithTolerance(valA.Float(), valB.Float(), 0.0001)
	case reflect.Slice:
		if valA.Len() != valB.Len() {
			return false
		}
		for i := 0; i < valA.Len(); i++ {
			if !deepEqual(valA.Index(i).Interface(), valB.Index(i).Interface()) {
				return false
			}
		}
		return true
	case reflect.Map:
		if valA.Len() != valB.Len() {
			return false
		}
		for _, key := range valA.MapKeys() {
			if !deepEqual(valA.MapIndex(key).Interface(), valB.MapIndex(key).Interface()) {
				return false
			}
		}
		return true
	case reflect.Interface, reflect.Pointer:
		if valA.Type() != valB.Type() {
			return false
		}
		return deepEqual(valA.Elem().Interface(), valB.Elem().Interface())
	case reflect.Struct:
		if valA.Type() != valB.Type() {
			return false
		}
		for i := 0; i < valA.NumField(); i++ {
			fieldA, fieldB := valA.Field(i), valB.Field(i)
			if !fieldA.CanInterface() {
				// field is not exported
				continue
			}
			if !deepEqual(fieldA.Interface(), fieldB.Interface()) {
				return false
			}
		}
		return true
	default:
		// I'm keeping it here for the sake of curiosity and learning about other cases.
		// If in rush, just enable reflect.DeepEqual instead.
		panic("not implemented")
		//return reflect.DeepEqual(a, b)
	}
}
